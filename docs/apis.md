# 接口

- `GET /chat/v1/rooms/{roomUID}` 获取房间信息
- `GET /chat/v1/rooms/{roomUID}/members` 列出成员
- `POST /chat/v1/rooms/{roomUID}/members` 创建成员（加入房间）
- `GET /chat/v1/rooms/{roomUID}/members/{memberUID}` 获取成员信息
- `DELETE /chat/v1/rooms/{roomUID}/members/{memberUID}` 删除成员（离开房间）
- `POST /chat/v1/rooms/{roomUID}/messages` 创建消息（发送消息）
- `GET /chat/v1/rooms/{roomUID}/messages` 监听消息
