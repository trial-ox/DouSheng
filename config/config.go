package config

var OneHour = 60 * 60
var OneDay = 60 * 60 * 24
var OneMinute = 60 * 1
var OneMonth = 60 * 60 * 24 * 30
var OneYear = 365 * 60 * 60 * 24

//jwt密钥
var Secret = "douyin"

// 每次视频流返回最大数
var VideoCount = 2

//评论操作状态：有效
const ValidComment = 0

//评论操作状态：取消
const InvalidComment = 1

const IsLike = 0     //点赞的状态
const Unlike = 1     //取消赞的状态
const LikeAction = 1 //点赞的行为

//格式化时间
const DateTime = "2023-02-21 15:04:05"

const DefaultRedisValue = -1 //redis中key对应的预设值，防脏读
