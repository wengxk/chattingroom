# chattingroom

基于go语言实现的一个windows控制台聊天室小应用

## TO-DO

1. map并发问题 【√已完成】
2. 心跳检测 【√已完成】

## 问题

描述：

&nbsp;&nbsp;&nbsp;&nbsp;当再开一个协程做心跳检测时，就有可能出现同一个tcp连接并发读写的问题，除非加锁，心跳时其他tcp操作阻塞，直到心跳完成，或是tcp有操作时，心跳阻塞，直到tcp操作完成。前期在设计服务器主动推送其他客户端发来的消息时，也是另外开了一个tcp连接，就是为了防止出现tcp并发读写，但是这样就维护两个tcp连接，一个是用于用户主动和服务器互动的连接，一个是专门用于接收服务器主动推送消息的连接，这无疑增加了连接成本和维护成本。

优化：

1. 只用一个tcp连接来完成所有操作
2. 解决一个tcp连接多个协程可能出现的并发读写问题

思路：

1. 所有请求都分配一个UID，并且该请求对应的响应也包含这个UID
2. 一个连接，两个协程，一个读，一个写
3. 建一个全局map，key为请求全局UID，value为期望值的管道
4. 在发请求时新建一个期望的map键值，添加到全局map中，此时该请求协程阻塞直到管道可以读取数据
5. 在读管道的协程里，根据请求响应返回的UID，在全局map中找出对应的管道并写入响应
6. 阻塞协程从管道中获取自己期望的数据，继续往下处理
