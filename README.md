# go-casper [![Travis](https://img.shields.io/travis/tcnksm/go-casper.svg?style=flat-square)][travis] [![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license] [![Go Documentation](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)][godocs]

[travis]: https://travis-ci.org/tcnksm/go-casper
[license]: https://github.com/tcnksm/go-casper/blob/master/LICENSE
[godocs]: http://godoc.org/github.com/tcnksm/go-casper

`go-casper` is golang implementation of [H2O](https://github.com/h2o/h2o)'s CASPer (cache-aware server-push). H2O is a server that provides full advantage of HTTP/2 features. 

The full documentation is available on [Godoc][godocs].

*NOTE*: The project is still under heavy implementation. The API may be changed in future and documentaion is incomplete.

## Example 

Below is a simple example of usage.

```golang
// Initialize casper with false-positive probability 
// 1/64 and number of assets 10.
pusher := casper.New(1<<6, 10)

http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {    
    
    // Execute cache aware server push.
    // It only push when client cookie indicate the assets are not cached.
    if _, err := pusher.Push(w, r, []string{"/static/example.jpg"}, nil); err != nil {
        log.Printf("[ERROR] Failed to push assets: %s", err)
    }

    // ...
})
```

You can find the complete example [here](/example).

## References

- http://blog.kazuhooku.com/2015/10/performance-of-http2-push-and-server.html
- https://github.com/h2o/h2o/issues/421
- https://h2o.examp1e.net/configure/http2_directives.html#http2-casper
- https://datatracker.ietf.org/doc/draft-kazuho-h2-cache-digest/
- http://www.slideshare.net/kazuho/cache-awareserverpush-in-h2o-version-15
- http://cdn.oreillystatic.com/en/assets/1/event/167/The%20promise%20of%20Push%20Presentation.pdf