# `AGAMOTTO`

<a href='https://github.com/jpoles1/gopherbadger' target='_blank'>![gopherbadger-tag-do-not-edit](https://img.shields.io/badge/Go%20Coverage-96%25-brightgreen.svg?longCache=true&style=flat)</a>

> Simple queue with delivery at a specific time in the future

```go
// Create new instance and in/out channels
// and start serving
s, in, out := agamotto.NewService(...)
go s.Serve()

// Listen to result from out channel in the goroutine
go func(){
    for entity := range out {
        fmt.Println(string(entity.data))
    }
}()

// Send the Event with 1 second time shift (for example)
in <- &agamotto.Event{
    When: uint64(time.Now().Unix()) + 1,
    Entity: &agamotto.Entity{
        data: []byte{"Hello, World!"},
    },
}
```
After 1 second you will be receive "Hello, World!" to the channel.