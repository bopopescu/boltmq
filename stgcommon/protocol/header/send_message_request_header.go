package header
// SendMessageRequestHeader: 发送消息请求头信息
// Author: yintongqiang
// Since:  2017/8/10

type SendMessageRequestHeader struct {
	ProducerGroup         string
	Topic                 string
	DefaultTopic          string
	DefaultTopicQueueNums int
	QueueId               int32
	SysFlag               int32
	BornTimestamp         int64
	Flag                  int
	Properties            string
	ReconsumeTimes        int32
	UnitMode              bool
}

func (header*SendMessageRequestHeader)CheckFields(){

}
