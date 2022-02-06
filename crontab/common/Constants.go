package common

//任务保存目录
const JOB_SAVE_DIR = "/cron/jobs/"  //任务保存目录
const JOB_KILL_DIR = "/cron/killer/" //任务强杀目录
const JOB_LOCK_DIR = "/cron/lock/" //任务锁目录


//保存任务事件
const JOB_SAVE_EVENT = 1
const JOB_DELETE_EVENT = 2
const JOB_KILL_EVENT = 3