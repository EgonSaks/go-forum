# Forum

To run the program...

Create dev.env to config folder in configs folder. 

Add there your google and github data as follows without "{}":
```
// Client ID
GOOGLE_KEY={YOUR CLIENT ID}

// Client secret
GOOGLE_SECRET={YOUR CLIENT ID}

// Client ID
GITHUB_KEY={YOUR CLIENT ID}

// Client secret
GITHUB_SECRET={YOUR CLIENT ID}
```

Then:

`go run cmd/web/*` 
or 
`sh dockerRun.sh`


```
GO_ENV=prod go run main.go

or  

export GO_ENV=prod
go run main.go
```

```
GO_ENV=dev go run main.go

or

export GO_ENV=dev
go run main.go
```

```
GO_ENV=test go run main.go

or

export GO_ENV=test
go run main.go
```