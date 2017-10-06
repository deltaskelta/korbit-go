### Korbit API Client in Go

```go
go get github.com/deltaskelta/korbit-go
```

```go
import korbit "github.com/deltaskelta/korbit-go"

func main() {
    korbit := korbit.NewKorbitAPI(APIKey, SecretKey, Username, Password)
    err := korbit.Login()
    if err != nil {
        panic(err)
    }

    // use the endpoints as needed...
}
```

### Contributing

In order to run the tests, you will need to define the login credentials somewhere in the
package. 
