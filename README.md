# Mercury

## Introduction
Mercury is an instant messaging server based on a go-micro implementation.

## About the name
The name was actually intended to be taken from Hermes, one of the twelve main Greek gods, but, you know, it was too expensive for me to match, so it was taken to correspond to Mercury in Roman mythology. He was Jupiter's most faithful messenger, delivering messages for Jupiter and completing the various tasks that Jupiter gave him. As the god of communication, he has superb wisdom and communication skills, and I think it is appropriate to use his name for instant messaging.

Translated with www.DeepL.com/Translator (free version) and a messenger app that is built on top of the protocol.


## How to use?
```shell script
$ ./mercury infra # provide configuration of each service
$ ./mercury logic 
$ ./mercury job 
$ ./mercury comet 
```

### Handshake
```json
{"operation": "handshake", "body": {"mid": "mid", "version": "v0.1", "user_agent": "user_agent", "device_id": "xxx", "token": "user_token"}}
```

### Connect
```json
{"operation": "connect", "body": {"mid": "mid", "token": "user_token"}}
```

### Push - single
```json
{"operation": "push", "body": {"mid": "mid", "message_type": "single", "receiver": "uid", "content_type": "text", "body": {"content": "Hello, World!"}, "mentions": []}}
```

### Push - group
```json
{"operation": "push", "body": {"mid": "mid", "message_type": "group", "receiver": "gid4Fl1QvXZpM4", "content_type": "text", "body": {"content": "Hello, World!"}, "mentions": ["uid7KA8fY5Jb3A"]}}
```

### Notification
```json
{"operation": "notification", "body": {"mid": "mid", "what": "keypress", "topic": "p2puN_f_2oWkUTsoDx9jklvcA"}}
```
