



## conc 的目标

- 更难出现goroutine泄漏
- 处理panic更友好
- 并发代码可读性高

## 主要封装功能 

- 对`WaitGroup`进行封装，避免了产生大量重复代码，并且也封装`recover`，安全性更高
- 提供`panics.Catcher`封装`recover`逻辑，统一捕获`panic`，打印调用栈一些信息
- 提供一个并发执行任务的`worker`池，可以控制并发度、`goroutine`可以进行复用，支持函数签名，同时提供了`stream`方法来保证结果有序
- 提供`ForEach`、`map`方法优雅的处理切片
- `conc.WatiGroup`对`sync.WaitGroup`进行了封装，对`Add`、`Done`、`Recover`进行了封装，提高了可读性，避免了冗余代码
- `pool`是一个并发的协程队列，可以控制协程的数量，实现上使用一个无缓冲的`channel`作为`worker`，如果`goroutine`执行速度快，避免了创建多个`goroutine`
- `stream`是一个保证顺序的并发协程队列，实现上使用`sync.Pool`在提交`goroutine`时控制顺序





