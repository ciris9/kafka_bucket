# kafka_bucket
施工中，
需求：
1. 需要消费kafka的大量消息，每秒消息数万级别起步
2. 可以集群部署
3. 处理的消息达到一定时间或一定大小就进行压缩，并上传到云存储桶中。
4. 需要保证消费kafka消息不丢失，主要体现在两个方面，一个是offset导致的消息丢失，另一个是宕机导致的缓冲区消息丢失。
