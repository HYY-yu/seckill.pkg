package elastic_job

import (
	"encoding/json"
)

type Job struct {
	Key       string                 // /system_name/job_id
	DelayTime int64                  // 延迟时间
	Cycle     bool                   // 是否周期循环
	Tag       string                 // Tag匹配Handler，无Tag的Job将不会被执行
	Args      map[string]interface{} // 任务参数
}

func (j *Job) MarshalJson() string {
	jJob, _ := json.Marshal(j)
	return string(jJob)
}

func UnmarshalJson(j string) (*Job, error) {
	var job Job
	err := json.Unmarshal([]byte(j), &job)
	return &job, err
}

type Handler func(j *Job) (err error)
