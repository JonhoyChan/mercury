/******************************************************************************
 *
 *  Description :
 *
 *    Topic initilization routines.
 *
 *****************************************************************************/

package session

import (
	"outgoing/app/gateway/chat/api"
	"outgoing/app/gateway/chat/stats"
	"outgoing/x"
	"outgoing/x/ecode"
	"outgoing/x/log"
	"outgoing/x/types"
	"strings"
	"time"
)

// topicInit reads an existing topic from database or creates a new topic
func topicInit(t *Topic, msg *sessionJoin, h *Hub) {
	timestamp := time.Now().UTC().Unix()

	var err error
	switch {
	case t.original == "me":
		// Request to load a 'me' topic. The topic always exists, the subscription is never new.
		err = initTopicMe(t, msg)
	//case t.original == "find":
	//	// Request to load a 'find' topic. The topic always exists, the subscription is never new.
	//	err = initTopicFind(t, join)
	case strings.HasPrefix(t.original, "uid") || strings.HasPrefix(t.original, "p2p"):
		// Request to load an existing or create a new p2p topic, then attach to it.
		err = initTopicP2P(t, msg)
	//case strings.HasPrefix(t.original, "new"):
	//	// Processing request to create a new group topic
	//	err = initTopicNewGrp(t, join)
	//case strings.HasPrefix(t.original, "grp"):
	//	// Load existing group topic
	//	err = initTopicGrp(t, join)
	//case t.original == "sys":
	//	// Initialize system topic.
	//	err = initTopicSys(t, join)
	default:
		// Unrecognized topic name
		err = ecode.ErrNotFound.ResetMessage("topic not found")
	}
	// Failed to create or load the topic.
	if err != nil {
		// Remove topic from cache to prevent hub from forwarding more messages to it.
		h.topicDel(msg.routeTo)

		log.Info("failed to load or create topic", log.Ctx{"route_to": msg.routeTo, "error": err})
		// TODO
		msg.session.queueOut(msg.session.serialize(nil, api.NewResponse(err, msg.mid, msg.original, timestamp)))

		// Re-queue pending requests to join the topic.
		for len(t.join) > 0 {
			join := <-t.join
			h.join <- join
		}

		// Reject all other pending requests
		for len(t.broadcast) > 0 {
			msg := <-t.broadcast
			if msg.Id != "" {
				msg.session.queueOut(ecode.ErrLocked(msg.Id, t.original, timestamp))
			}
		}
		for len(t.leave) > 0 {
			msg := <-t.leave
			if msg.data != nil {
				msg.session.queueOut(api.NewResponse(ecode.ErrLocked, msg.mid, t.original, timestamp))
			}
		}

		if len(t.exit) > 0 {
			msg := <-t.exit
			msg.done <- true
		}

		return
	}

	t.computePerUserAcsUnion()

	// prevent newly initialized topics to go live while shutdown in progress
	//if globals.shuttingDown {
	//	h.topicDel(msg.routeTo)
	//	return
	//}

	if t.isDeleted() {
		// Someone deleted the topic while we were trying to create it.
		return
	}

	stats.Set("LiveTopics", 1, true)
	stats.Set("TotalTopics", 1, true)

	// Topic will check access rights, send invite to p2p user, send {ctrl} message to the initiator session
	if msg.pkt.Sub != nil {
		t.join <- join
	}

	t.markPaused(false)
	//if t.cat == types.TopicCatFnd || t.cat == types.TopicCatSys {
	//	t.markLoaded()
	//}

	go t.run(h)
}

// Initialize 'me' topic.
func initTopicMe(t *Topic, msg *sessionJoin) error {
	t.category = types.TopicCategoryMe

	// TODO
	user, err := store.Users.Get(types.ParseUserId(t.name))
	if err != nil {
		// Log out the session
		msg.session.uid = types.ZeroUid
		return err
	} else if user == nil {
		// Log out the session
		msg.session.uid = types.ZeroUid
		return ecode.ErrUserNotFound
	}

	// User's default access for p2p topics
	//t.accessAuth = user.Access.Auth
	//t.accessAnon = user.Access.Anon

	if err = t.loadSubscribers(); err != nil {
		return err
	}

	t.createdAt = user.CreatedAt
	t.updatedAt = user.UpdatedAt

	// The following values are exlicitly not set for 'me'.
	// t.touched, t.lastId, t.delId
	// 'me' has no owner, t.owner = nil

	// Initiate User Agent with the UA of the creating session to report it later
	t.userAgent = msg.session.userAgent
	// Initialize channel for receiving user agent and session online updates.
	//t.supd = make(chan *sessionUpdate, 32)
	// Allocate storage for contacts.
	t.perSubs = make(map[string]perSubsData)

	return nil
}

// Initialize 'find' topic
//func initTopicFind(t *Topic, msg *sessionJoin) error {
//	t.category = types.TopicCategoryFind
//
//	uid := types.ParseUserId(msg.pkt.AsUser)
//	if uid.IsZero() {
//		return types.ErrNotFound
//	}
//
//	user, err := store.Users.Get(uid)
//	if err != nil {
//		return err
//	} else if user == nil {
//		msg.session.uid = types.ZeroUid
//		return ecode.ErrUserNotFound
//	}
//
//	// Make sure no one can join the topic.
//	//t.accessAuth = getDefaultAccess(t.category, true)
//	//t.accessAnon = getDefaultAccess(t.category, false)
//
//	if err = t.loadSubscribers(); err != nil {
//		return err
//	}
//
//	t.createdAt = user.CreatedAt
//	t.updatedAt = user.UpdatedAt
//
//	// 'find' has no owner, t.owner = nil
//
//	// Publishing to find is not supported
//	// t.lastId = 0, t.delId = 0, t.touched = nil
//
//	return nil
//}

// Load or create a P2P topic.
// There is a reace condition when two users try to create a p2p topic at the same time.
func initTopicP2P(t *Topic, msg *sessionJoin) error {
	// Handle the following cases:
	// 1. Neither topic nor subscriptions exist: create a new p2p topic & subscriptions.
	// 2. Topic exists, one of the subscriptions is missing:
	// 2.1 Requester's subscription is missing, recreate it.
	// 2.2 Other user's subscription is missing, treat like a new request for user 2.
	// 3. Topic exists, both subscriptions are missing: should not happen, fail.
	// 4. Topic and both subscriptions exist: attach to topic

	t.category = types.TopicCategoryP2P

	// Check if the topic already exists
	topic, err := store.Topics.Get(t.name)
	if err != nil {
		return err
	}

	// If topic exists, load subscriptions
	var subs []types.Subscription
	if topic != nil {
		// Subs already have Public swapped
		if subs, err = store.Topics.GetUsers(t.name, nil); err != nil {
			return err
		}

		// Case 3, fail
		if len(subs) == 0 {
			log.Warn(x.Sprintf("missing both subscriptions for %s (SHOULD NEVER HAPPEN!)", t.name))
			return ecode.ErrInternalServer
		}

		t.createdAt = topic.CreatedAt
		t.updatedAt = topic.UpdatedAt
		if topic.TouchedAt != 0 {
			t.touchedAt = topic.TouchedAt
		}
		t.lastID = topic.SeqId
		t.delID = topic.DelId
	}

	// t.owner is blank for p2p topics

	// Default user access to P2P topics is not set because it's unused.
	// Other users cannot join the topic because of how topic name is constructed.
	// The two participants set each other's access instead.
	// t.accessAuth = getDefaultAccess(t.cat, true)
	// t.accessAnon = getDefaultAccess(t.cat, false)

	// t.public is not used for p2p topics since each user get a different public

	if topic != nil && len(subs) == 2 {
		// Case 4.
		for i := 0; i < 2; i++ {
			uid := types.ParseUid(subs[i].User)
			t.perUser[uid] = perUserData{
				topicName: types.ParseUid(subs[(i+1)%2].User).UID(),
				delID:     subs[i].DelId,
				recvID:    subs[i].RecvSeqId,
				readID:    subs[i].ReadSeqId,
			}
		}
	} else {
		// Cases 1 (new topic), 2 (one of the two subscriptions is missing: either it's a new request
		// or the subscription was deleted)
		var userData perUserData

		// Fetching records for both users.
		// Requester.
		uid1 := msg.session.uid
		// The other user.
		uid2 := types.ParseUid(t.original)
		// User index: u1 - requester, u2 - responder, the other user

		users, err := store.Users.GetAll(uid1, uid2)
		if err != nil {
			return err
		}
		if len(users) != 2 {
			// Invited user does not exist
			return ecode.ErrUserNotFound
		}

		// Figure out which subscriptions are missing: User1's, User2's or both.
		var sub1, sub2 *types.Subscription
		// Set to true if only requester's subscription has to be created.
		var user1only bool
		if len(subs) == 1 {
			if subs[0].User == uid1.UID() {
				// User2's subscription is missing, user1's exists
				sub1 = &subs[0]
			} else {
				// User1's is missing, user2's exists
				sub2 = &subs[0]
				user1only = true
			}
		}

		// Other user's (responder's) subscription is missing
		if sub2 == nil {
			sub2 = &types.Subscription{
				User:  uid2.UID(),
				Topic: t.name,
			}

			// Mark the entire topic as new.
			pktsub.Created = true
		}

		// Requester's subscription is missing:
		// a. requester is starting a new topic
		// b. requester's subscription is missing: deleted or creation failed
		if sub1 == nil {
			sub1 = &types.Subscription{
				User:  uid1.UID(),
				Topic: t.name,
			}

			// Mark this subscription as new
			pktsub.Newsub = true
		}

		// Create everything
		if topic == nil {
			if err = store.Topics.CreateP2P(sub1, sub2); err != nil {
				return err
			}

			t.createdAt = sub1.CreatedAt.Unix()
			t.updatedAt = sub1.UpdatedAt.Unix()
			t.touchedAt = t.updatedAt

			// t.lastId is not set (default 0) for new topics
		} else {
			// TODO possibly update subscription, if changed

			// Recreate one of the subscriptions
			var subToMake *types.Subscription
			if user1only {
				subToMake = sub1
			} else {
				subToMake = sub2
			}
			if err = store.Subs.Create(subToMake); err != nil {
				return err
			}
		}

		userData.topicName = uid2.UID()
		userData.delID = sub1.DelId
		userData.readID = sub1.ReadSeqId
		userData.recvID = sub1.RecvSeqId
		t.perUser[uid1] = userData

		t.perUser[uid2] = perUserData{
			topicName: uid1.UID(),
			delID:     sub2.DelId,
			readID:    sub2.ReadSeqId,
			recvID:    sub2.RecvSeqId,
		}
	}

	// Clear original topic name.
	t.original = ""

	return nil
}
