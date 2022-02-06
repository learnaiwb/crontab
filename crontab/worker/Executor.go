package worker

import (
	"github.com/learnaiwb/crontab/crontab/common"
	"math/rand"
	"os/exec"
	"time"
)

//任务执行器
type Executor struct {

}

var (
	G_executor *Executor
)
//执行一个任务
func (executor *Executor) ExecutorJob(info *common.JobExecuteInfo)  {
	go func() {
		var (
			cmd *exec.Cmd
			out []byte
			res *common.JobExecuteResult
			err error
			jobLock *JobLock
		)

		res = &common.JobExecuteResult{
			ExecuteInfo: info,
			Output:      nil,
			Err:         nil,
			StartTime:   time.Now(),
			EndTime:     time.Time{},
		}

		//初始化分布式锁
		jobLock = G_jobMgr.CreateJobLock(info.Job.Name)
		//随机睡眠0-1s 上锁均衡
		time.Sleep(time.Duration(rand.Intn(1000))*time.Millisecond)

		//上锁
		err = jobLock.TryLock()
		defer jobLock.UnLock()
		if err != nil {
			res.Err = err
			res.EndTime = time.Now()
		}else {
			res.StartTime = time.Now()
			//执行shell命令
			cmd = exec.CommandContext(info.Ctx,"bash.exe","-c",info.Job.Command)
			//执行并捕获输出
			out,err = cmd.CombinedOutput()
			//任务执行完毕 把执行结果返回给scheduler ,scheduler会从executingTable中删除记录
			res.Output = out
			res.Err = err
			res.EndTime = time.Now()

		}
		G_scheduler.PushJobResult(res)
	}()
}

func InitExecutor() (err error) {
	G_executor = &Executor{}
	return
}
