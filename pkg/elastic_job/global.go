package elastic_job

import (
	"sync"
)

var once sync.Once

var ej ElasticJob

func InitGlobalJob(opts ...Options) (err error) {
	once.Do(func() {
		e, er := New(opts...)
		if er != nil {
			err = er
		}
		ej = e
	})
	return
}

func Get() ElasticJob {
	if ej == nil {
		panic("you must run InitGlobalJob first. ")
	}
	return ej
}
