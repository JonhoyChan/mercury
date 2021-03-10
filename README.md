# Mercury

Mercury is an instant messaging backend service


### Handshake
{"operation": "handshake", "body": {"mid": "mid", "version": "v0.1", "user_agent": "user_agent", "device_id": "xxx", "token": "user_token"}}

### Connect
{"operation": "connect", "body": {"mid": "mid", "token": "user_token"}}

### Push - single
{"operation": "push", "body": {"mid": "mid", "message_type": "single", "receiver": "uid", "content_type": "text", "body": {"content": "Hello, World!"}, "mentions": []}}

### Push - group
{"operation": "push", "body": {"mid": "mid", "message_type": "group", "receiver": "gid4Fl1QvXZpM4", "content_type": "text", "body": {"content": "Hello, World!"}, "mentions": ["uid7KA8fY5Jb3A"]}}

### Notification
{"operation": "notification", "body": {"mid": "mid", "what": "keypress", "topic": "p2puN_f_2oWkUTsoDx9jklvcA"}}
