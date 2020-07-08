package session

import (
	"log"
	"outgoing/x/types"
	"time"
)

// Topic is an isolated communication channel
type Topic struct {
	// Еxpanded/unique name of the topic.
	name string
	// For single-user topics session-specific topic name, such as 'me', otherwise the same as 'name'.
	original string

	// Topic category
	category types.TopicCategory

	// Timestamp when the topic was first created.
	createdAt int64
	// Timestamp when the topic was last updated.
	updatedAt int64
	// Timestamp of the last outgoing message.
	touchedAt int64

	// Server-side ID of the last data message
	lastID int
	// ID of the deletion operation. Not an ID of the message.
	delID int

	// Last published userAgent ('me' topic only)
	userAgent string

	// User ID of the topic owner/creator. Could be zero.
	owner types.Uid

	// Default access mode
	//accessAuth types.AccessMode
	//accessAnon types.AccessMode

	// Topic's per-subscriber data
	perUser map[types.Uid]perUserData

	// User's contact list (not nil for 'me' topic only).
	// The map keys are UserIds for P2P topics and grpXXX for group topics.
	perSubs map[string]perSubsData

	// Sessions attached to this topic. The UID kept here may not match Session.uid if session is
	// subscribed on behalf of another user.
	sessions map[*Session]perSessionData

	// Inbound {data} and {pres} messages from sessions or other topics, already converted to SCM. Buffered = 256
	broadcast chan []byte
	// Subscribe requests from sessions, buffered = 32
	join chan *sessionJoin
	// Unsubscribe requests from sessions, buffered = 32
	leave chan *sessionLeave
	// Session updates: background sessions coming online, User Agent changes. Buffered = 32
	//supd chan *sessionUpdate
	// Channel to terminate topic  -- either the topic is deleted or system is being shut down. Buffered = 1.
	exit chan *shutDown

	// Flag which tells topic lifecycle status: new, ready, paused, marked for deletion.
	status int32
}

// Holds topic's cache of per-subscriber data
type perUserData struct {
	// Timestamps when the subscription was created and updated
	createdAt int64
	updatedAt int64

	// Count of subscription online and announced (presence not deferred).
	online int

	// Last t.lastId reported by user through {pres} as received or read
	recvID int
	readID int
	// ID of the latest Delete operation
	delID int

	//modeWant  types.AccessMode
	//modeGiven types.AccessMode

	// P2P only:
	topicName string
	deleted   bool
}

// Holds user's (on 'me' topic) cache of subscription data
type perSubsData struct {
	// The other user's/topic's online status as seen by this user.
	online bool
	// True if we care about the updates from the other user/topic: (want&given).IsPresencer().
	// Does not affect sending notifications from this user to other users.
	enabled bool
}

// Data related to a subscription of a session to a topic.
type perSessionData struct {
	// ID of the subscribed user (asUid); not necessarily the session owner.
	// Could be zero for multiplexed sessions in cluster.
	uid types.Uid
}

// Reasons why topic is being shut down.
const (
	// StopNone no reason given/default.
	StopNone = iota
	// StopShutdown terminated due to system shutdown.
	StopShutdown
	// StopDeleted terminated due to being deleted.
	StopDeleted
)

// Topic shutdown
type shutDown struct {
	// Channel to report back completion of topic shutdown. Could be nil
	done chan<- bool
	// Topic is being deleted as opposite to total system shutdown
	reason int
}

func (t *Topic) run(hub *Hub) {
	// Kills topic after a period of inactivity.
	keepAlive := idleMasterTopicTimeout
	killTimer := time.NewTimer(time.Hour)
	killTimer.Stop()

	// Notifies about user agent change. 'me' only
	uaTimer := time.NewTimer(time.Minute)
	var currentUserAgent string
	uaTimer.Stop()

	// Ticker for deferred presence notifications.
	deferredNotifyTimer := time.NewTimer(time.Millisecond * 500)

	for {
		select {
		case join := <-t.join:
			// Request to add a connection to this topic
			if t.isInactive() {
				join.session.queueOut(ErrLocked(join.mid, t.original, time.Now().UTC().Unix()))
			} else {
				// The topic is alive, so stop the kill timer, if it's ticking. We don't want the topic to die
				// while processing the call
				killTimer.Stop()
				if err := t.handleSubscription(hub, join); err == nil {
					if join.pkt.Sub.Created {
						// Call plugins with the new topic
						pluginTopic(t, plgActCreate)
					}
				} else {
					if len(t.sessions) == 0 && t.cat != types.TopicCatSys {
						// Failed to subscribe, the topic is still inactive
						killTimer.Reset(keepAlive)
					}
					log.Printf("topic[%s] subscription failed %v, sid=%s", t.name, err, join.sess.sid)
				}
			}
		case leave := <-t.unreg:
			// Remove connection from topic; session may continue to function
			now := types.TimeNow()

			// userId.IsZero() == true when the entire session is being dropped.
			var asUid types.Uid
			if leave.pkt != nil {
				asUid = types.ParseUserId(leave.pkt.AsUser)
			}

			if t.isInactive() {
				if !asUid.IsZero() && leave.pkt != nil {
					leave.sess.queueOut(ErrLocked(leave.pkt.Id, t.original(asUid), now))
				}
				continue

			} else if leave.pkt != nil && leave.pkt.Leave.Unsub {
				// User wants to leave and unsubscribe.
				// asUid must not be Zero.
				if err := t.replyLeaveUnsub(hub, leave.sess, asUid, leave.pkt.Id); err != nil {
					log.Println("failed to unsub", err, leave.sess.sid)
					continue
				}
			} else if pssd, _ := t.remSession(leave.sess, asUid); pssd != nil {
				// Just leaving the topic without unsubscribing.

				var uid types.Uid
				if leave.sess.isProxy() {
					// Multiplexing session, multiple UIDs.
					uid = asUid
				} else {
					// Simple session, single UID.
					uid = pssd.uid
				}

				var pud perUserData
				// uid may be zero when a proxy session is trying to terminate (it called unsubAll).
				if !uid.IsZero() {
					// UID not zero: one user removed.
					pud = t.perUser[uid]
					if !leave.sess.background {
						pud.online--
					}
				} else if len(pssd.muids) > 0 {
					// UID is zero: multiplexing session is dropped altogether.
					// Using new 'uid' and 'pud' variables.
					for _, uid := range pssd.muids {
						pud := t.perUser[uid]
						pud.online--
						t.perUser[uid] = pud
					}
				} else if !leave.sess.isCluster() {
					log.Panic("cannot determine uid: leave req=", leave)
				}

				switch t.cat {
				case types.TopicCatMe:
					mrs := t.mostRecentSession()
					if mrs == nil {
						// Last session
						mrs = leave.sess
					} else {
						// Change UA to the most recent live session and announce it. Don't block.
						select {
						case t.supd <- &sessionUpdate{userAgent: mrs.userAgent}:
						default:
						}
					}

					meUid := uid
					if meUid.IsZero() {
						// The entire multiplexing session is being dropped. Need to find owner's UID.
						// May panic only if pssd.muids is empty, but it should not be empty at this point.
						meUid = pssd.muids[0]
					}
					// Update user's last online timestamp & user agent. Only one user can be subscribed to 'me' topic.
					if err := store.Users.UpdateLastSeen(meUid, mrs.userAgent, now); err != nil {
						log.Println(err)
					}
				case types.TopicCatFnd:
					// FIXME: this does not work correctly in case of a multiplexing query.
					// Remove ephemeral query.
					t.fndRemovePublic(leave.sess)
				case types.TopicCatGrp:
					// Topic is going offline: notify online subscribers on 'me'.
					readFilter := &presFilters{filterIn: types.ModeRead}
					if !uid.IsZero() {
						if pud.online == 0 {
							t.presSubsOnline("off", uid.UserId(), nilPresParams, readFilter, "")
						}
					} else if len(pssd.muids) > 0 {
						for _, uid := range pssd.muids {
							if t.perUser[uid].online == 0 {
								t.presSubsOnline("off", uid.UserId(), nilPresParams, readFilter, "")
							}
						}
					}
				}

				if !uid.IsZero() {
					t.perUser[uid] = pud

					// Respond if contains an id.
					if leave.pkt != nil {
						leave.sess.queueOut(NoErr(leave.pkt.Id, t.original(uid), now))
					}
				}
			}

			// If there are no more subscriptions to this topic, start a kill timer
			if len(t.sessions) == 0 && t.cat != types.TopicCatSys {
				killTimer.Reset(keepAlive)
			}

		case msg := <-t.broadcast:
			// Content message intended for broadcasting to recipients
			t.handleBroadcast(msg)

		case meta := <-t.meta:
			// Request to get/set topic metadata
			asUid := types.ParseUserId(meta.pkt.AsUser)
			authLevel := auth.Level(meta.pkt.AuthLvl)
			switch {
			case meta.pkt.Get != nil:
				// Get request
				if meta.pkt.MetaWhat&constMsgMetaDesc != 0 {
					if err := t.replyGetDesc(meta.sess, asUid, meta.pkt.Get.Id, meta.pkt.Get.Desc); err != nil {
						log.Printf("topic[%s] meta.Get.Desc failed: %s", t.name, err)
					}
				}
				if meta.pkt.MetaWhat&constMsgMetaSub != 0 {
					if err := t.replyGetSub(meta.sess, asUid, authLevel, meta.pkt.Get.Id, meta.pkt.Get.Sub); err != nil {
						log.Printf("topic[%s] meta.Get.Sub failed: %s", t.name, err)
					}
				}
				if meta.pkt.MetaWhat&constMsgMetaData != 0 {
					if err := t.replyGetData(meta.sess, asUid, meta.pkt.Get.Id, meta.pkt.Get.Data); err != nil {
						log.Printf("topic[%s] meta.Get.Data failed: %s", t.name, err)
					}
				}
				if meta.pkt.MetaWhat&constMsgMetaDel != 0 {
					if err := t.replyGetDel(meta.sess, asUid, meta.pkt.Get.Id, meta.pkt.Get.Del); err != nil {
						log.Printf("topic[%s] meta.Get.Del failed: %s", t.name, err)
					}
				}
				if meta.pkt.MetaWhat&constMsgMetaTags != 0 {
					if err := t.replyGetTags(meta.sess, asUid, meta.pkt.Get.Id); err != nil {
						log.Printf("topic[%s] meta.Get.Tags failed: %s", t.name, err)
					}
				}
				if meta.pkt.MetaWhat&constMsgMetaCred != 0 {
					log.Printf("topic[%s] handle getCred", t.name)
					if err := t.replyGetCreds(meta.sess, asUid, meta.pkt.Get.Id); err != nil {
						log.Printf("topic[%s] meta.Get.Creds failed: %s", t.name, err)
					}
				}

			case meta.pkt.Set != nil:
				// Set request
				if meta.pkt.MetaWhat&constMsgMetaDesc != 0 {
					if err := t.replySetDesc(meta.sess, asUid, meta.pkt.Set); err == nil {
						// Notify plugins of the update
						pluginTopic(t, plgActUpd)
					} else {
						log.Printf("topic[%s] meta.Set.Desc failed: %v", t.name, err)
					}
				}
				if meta.pkt.MetaWhat&constMsgMetaSub != 0 {
					if err := t.replySetSub(hub, meta.sess, meta.pkt); err != nil {
						log.Printf("topic[%s] meta.Set.Sub failed: %v", t.name, err)
					}
				}
				if meta.pkt.MetaWhat&constMsgMetaTags != 0 {
					if err := t.replySetTags(meta.sess, asUid, meta.pkt.Set); err != nil {
						log.Printf("topic[%s] meta.Set.Tags failed: %v", t.name, err)
					}
				}
				if meta.pkt.MetaWhat&constMsgMetaCred != 0 {
					if err := t.replySetCred(meta.sess, asUid, authLevel, meta.pkt.Set); err != nil {
						log.Printf("topic[%s] meta.Set.Cred failed: %v", t.name, err)
					}
				}

			case meta.pkt.Del != nil:
				// Del request
				var err error
				switch meta.pkt.MetaWhat {
				case constMsgDelMsg:
					err = t.replyDelMsg(meta.sess, asUid, meta.pkt.Del)
				case constMsgDelSub:
					err = t.replyDelSub(hub, meta.sess, asUid, meta.pkt.Del)
				case constMsgDelTopic:
					err = t.replyDelTopic(hub, meta.sess, asUid, meta.pkt.Del)
				case constMsgDelCred:
					err = t.replyDelCred(hub, meta.sess, asUid, authLevel, meta.pkt.Del)
				}

				if err != nil {
					log.Printf("topic[%s] meta.Del failed: %v", t.name, err)
				}
			}
		case upd := <-t.supd:
			if upd.sess != nil {
				// 'me' & 'grp' only. Background session timed out and came online.
				t.sessToForeground(upd.sess)
			} else if currentUA != upd.userAgent {
				if t.cat != types.TopicCatMe {
					log.Panicln("invalid topic category in UA update", t.name)
				}
				// 'me' only. Process an update to user agent from one of the sessions.
				currentUA = upd.userAgent
				uaTimer.Reset(uaTimerDelay)
			}

		case <-uaTimer.C:
			// Publish user agent changes after a delay
			if currentUA == "" || currentUA == t.userAgent {
				continue
			}
			t.userAgent = currentUA
			t.presUsersOfInterest("ua", t.userAgent)

		case <-killTimer.C:
			// Topic timeout
			hub.unreg <- &topicUnreg{rcptTo: t.name}
			defrNotifTimer.Stop()
			if t.cat == types.TopicCatMe {
				uaTimer.Stop()
				t.presUsersOfInterest("off", currentUA)
			} else if t.cat == types.TopicCatGrp {
				t.presSubsOffline("off", nilPresParams, nilPresFilters, nilPresFilters, "", false)
			}

		case sd := <-t.exit:
			// Handle four cases:
			// 1. Topic is shutting down by timer due to inactivity (reason == StopNone)
			// 2. Topic is being deleted (reason == StopDeleted)
			// 3. System shutdown (reason == StopShutdown, done != nil).
			// 4. Cluster rehashing (reason == StopRehashing)

			if sd.reason == StopDeleted {
				if t.cat == types.TopicCatGrp {
					t.presSubsOffline("gone", nilPresParams, nilPresFilters, nilPresFilters, "", false)
				}
				// P2P users get "off+remove" earlier in the process

				// Inform plugins that the topic is deleted
				pluginTopic(t, plgActDel)

			} else if sd.reason == StopRehashing {
				// Must send individual messages to sessions because normal sending through the topic's
				// broadcast channel won't work - it will be shut down too soon.
				t.presSubsOnlineDirect("term", nilPresParams, nilPresFilters, "")
			}
			// In case of a system shutdown don't bother with notifications. They won't be delivered anyway.

			// Tell sessions to remove the topic
			for s := range t.sessions {
				s.detach <- t.name
			}

			usersRegisterTopic(t, false)

			// Report completion back to sender, if 'done' is not nil.
			if sd.done != nil {
				sd.done <- true
			}
			return
		}
	}
}
