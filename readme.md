transcription-service/
├── cmd/
│   └── server/
│       └── main.go         # Entry point
├── internal/
│   ├── auth/              # Authentication logic
│   │   └── auth.go
│   ├── storage/           # S3 operations
│   │   └── storage.go
│   ├── api/              # HTTP handlers
│   │   └── handlers.go
│   └── models/           # Data structures
│       └── models.go
├── config/
│   └── config.go         # Configuration
└── go.mod               # Dependencies file

#added \r or \n instead of \r\n so fixed.
