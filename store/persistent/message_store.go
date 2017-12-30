// Copyright 2017 luoji

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//    http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package persistent

import (
	"sync"

	"github.com/boltmq/boltmq/store/stats"
)

type consumeQueueTable struct {
	consumeQueues   map[int32]*consumeQueue
	consumeQueuesMu sync.RWMutex
}

func newConsumeQueueTable() *consumeQueueTable {
	table := new(consumeQueueTable)
	//table.consumeQueues = make(map[int32]*consumeQueue)
	return table
}

// PersistentMessageStore 存储层对外提供的接口
// Author zhoufei
// Since 2017/9/6
type PersistentMessageStore struct {
	config               *Config // 存储配置
	clog                 *commitLog
	consumeTopicTable    map[string]*consumeQueueTable
	consumeQueueTableMu  sync.RWMutex
	cleanCLogService     *cleanCommitLogService     // 清理物理文件服务
	dispatchMsgService   *dispatchMessageService    // 分发消息索引服务
	allocateMFileService *allocateMappedFileService // 预分配文件
	idxService           *indexService              // 消息索引服务
	scheduleMsgService   *scheduleMessageService    // 定时服务
	runFlags             *runningFlags              // 运行过程标志位
	steCheckpoint        *storeCheckpoint
	storeStats           stats.StoreStats // 运行时数据统计
	/*
		MessageFilter            *DefaultMessageFilter // 消息过滤
		//MessageStoreConfig       *MessageStoreConfig   // 存储配置
		//CommitLog                *CommitLog
		//consumeTopicTable        map[string]*consumeQueueTable
		//consumeQueueTableMu      *sync.RWMutex
		FlushConsumeQueueService *FlushConsumeQueueService // 逻辑队列刷盘服务
		//CleanCommitLogService    *CleanCommitLogService    // 清理物理文件服务
		CleanConsumeQueueService *CleanConsumeQueueService // 清理逻辑文件服务
		//DispatchMessageService   *DispatchMessageService   // 分发消息索引服务
		//IndexService             *IndexService             // 消息索引服务
		//AllocateMapedFileService *AllocateMapedFileService // 从物理队列解析消息重新发送到逻辑队列
		ReputMessageService      *ReputMessageService      // 从物理队列解析消息重新发送到逻辑队列
		HAService                *HAService                // HA服务
		//ScheduleMessageService   *ScheduleMessageService   // 定时服务
		TransactionStateService  *TransactionStateService  // 分布式事务服务
		TransactionCheckExecuter *TransactionCheckExecuter // 事务回查接口
		StoreStatsService        *StoreStatsService        // 运行时数据统计
		//RunningFlags             *RunningFlags             // 运行过程标志位
		SystemClock              *stgcommon.SystemClock    // 优化获取时间性能，精度1ms
		ShutdownFlag             bool                      // 存储服务是否启动
		//StoreCheckpoint          *StoreCheckpoint
		BrokerStatsManager       *stats.BrokerStatsManager
		storeTicker              *timeutil.Ticker
		printTimes               int64
	*/
}

// GetMaxOffsetInQueue 获取指定队列最大Offset 如果队列不存在，返回-1
// Author: zhoufei, <zhoufei17@gome.com.cn>
// Since: 2017/9/20
func (ms *PersistentMessageStore) GetMaxOffsetInQueue(topic string, queueId int32) int64 {
	logic := ms.findConsumeQueue(topic, queueId)
	if logic != nil {
		return logic.getMaxOffsetInQueue()
	}

	return -1
}

func (ms *PersistentMessageStore) findConsumeQueue(topic string, queueId int32) *consumeQueue {
	ms.consumeQueueTableMu.RLock()
	cqMap, ok := ms.consumeTopicTable[topic]
	ms.consumeQueueTableMu.RUnlock()

	if !ok {
		ms.consumeQueueTableMu.Lock()
		cqMap = newConsumeQueueTable()
		ms.consumeTopicTable[topic] = cqMap
		ms.consumeQueueTableMu.Unlock()
	}

	cqMap.consumeQueuesMu.RLock()
	logic, ok := cqMap.consumeQueues[queueId]
	cqMap.consumeQueuesMu.RUnlock()

	if !ok {
		storePathRootDir := getStorePathConsumeQueue(ms.config.StorePathRootDir)
		cqMap.consumeQueuesMu.Lock()
		logic = newConsumeQueue(topic, queueId, storePathRootDir, int64(ms.config.getMappedFileSizeConsumeQueue()), ms)
		cqMap.consumeQueues[queueId] = logic
		cqMap.consumeQueuesMu.Unlock()
	}

	return logic
}

func (ms *PersistentMessageStore) putDispatchRequest(dRequest *dispatchRequest) {
	ms.dispatchMsgService.putRequest(dRequest)
}

func (ms *PersistentMessageStore) truncateDirtyLogicFiles(phyOffset int64) {
	for _, queueMap := range ms.consumeTopicTable {
		for _, logic := range queueMap.consumeQueues {
			logic.truncateDirtyLogicFiles(phyOffset)
		}
	}
}

func (ms *PersistentMessageStore) destroyLogics() {
	for _, queueMap := range ms.consumeTopicTable {
		for _, logic := range queueMap.consumeQueues {
			logic.destroy()
		}
	}
}

func (ms *PersistentMessageStore) putMessagePostionInfo(topic string, queueId int32, offset int64, size int64,
	tagsCode, storeTimestamp, logicOffset int64) {
	cq := ms.findConsumeQueue(topic, queueId)
	if cq != nil {
		cq.putMessagePostionInfoWrapper(offset, size, tagsCode, storeTimestamp, logicOffset)
	}
}
