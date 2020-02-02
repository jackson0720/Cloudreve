package task

import (
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/util"
)

// 任务类型
const (
	// CompressTaskType 压缩任务
	CompressTaskType = iota
)

// 任务状态
const (
	// Queued 排队中
	Queued = iota
	// Processing 处理中
	Processing
	// Error 失败
	Error
	// Canceled 取消
	Canceled
	// Complete 完成
	Complete
)

// 任务进度
const (
	// Compressing 压缩中
	CompressingProgress = iota
	// Decompressing 解压缩中
	DecompressingProgress
	// Downloading 下载中
	DownloadingProgress
	// Transferring 转存中
	TransferringProgress
)

// Job 任务接口
type Job interface {
	Type() int           // 返回任务类型
	Creator() uint       // 返回创建者ID
	Props() string       // 返回序列化后的任务属性
	Model() *model.Task  // 返回对应的数据库模型
	SetStatus(int)       // 设定任务状态
	Do()                 // 开始执行任务
	SetError(*JobError)  // 设定任务失败信息
	GetError() *JobError // 获取任务执行结果，返回nil表示成功完成执行
}

// JobError 任务失败信息
type JobError struct {
	Msg string
}

// Record 将任务记录到数据库中
func Record(job Job) (*model.Task, error) {
	record := model.Task{
		Status:   Queued,
		Type:     job.Type(),
		UserID:   job.Creator(),
		Progress: 0,
		Error:    "",
		Props:    job.Props(),
	}
	_, err := record.Create()
	return &record, err
}

// Resume 从数据库中恢复未完成任务
func Resume() {
	tasks := model.GetTasksByStatus(Queued)
	if len(tasks) == 0 {
		return
	}
	util.Log().Info("从数据库中恢复 %d 个未完成任务", len(tasks))

	for i := 0; i < len(tasks); i++ {
		job, err := GetJobFromModel(&tasks[i])
		if err != nil {
			util.Log().Warning("无法恢复任务，%s", err)
			continue
		}

		TaskPoll.Submit(job)
	}
}

// GetJobFromModel 从数据库给定模型获取任务
func GetJobFromModel(task *model.Task) (Job, error) {
	switch task.Type {
	case CompressTaskType:
		return NewCompressTaskFromModel(task)
	default:
		return nil, ErrUnknownTaskType
	}
}