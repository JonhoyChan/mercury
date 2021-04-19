# Mercury

Mercury is an instant messaging backend service


# How to use?
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
