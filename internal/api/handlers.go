package api

import (
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"

    "github.com/gorilla/mux"
    "github.com/supertokens/supertokens-golang/recipe/passwordless"
    "github.com/supertokens/supertokens-golang/recipe/passwordless/plessmodels"
    "github.com/supertokens/supertokens-golang/recipe/session"
    "github.com/supertokens/supertokens-golang/recipe/session/sessmodels"
    "github.com/supertokens/supertokens-golang/supertokens"
)

const (
    EC2_PUBLIC_IP = "3.131.82.143"
    PORT          = "8080"
)

func init() {
    logFile, err := os.OpenFile("logs/auth.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatal(err)
    }
    mw := io.MultiWriter(os.Stdout, logFile)
    log.SetOutput(mw)
    log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func InitSuperTokens() error {
    log.Println("Starting SuperTokens initialization...")
    
    apiDomain := "http://" + EC2_PUBLIC_IP + ":" + PORT
    apiBasePath := "/auth"
    websiteBasePath := "/auth"

    err := supertokens.Init(supertokens.TypeInput{
        Supertokens: &supertokens.ConnectionInfo{
            ConnectionURI: "http://localhost:3567",
            APIKey:        "supertokens-long-api-key-123456789",
        },
        AppInfo: supertokens.AppInfo{
            AppName:         "Transcription Service",
            APIDomain:       apiDomain,
            WebsiteDomain:   apiDomain,
            APIBasePath:     &apiBasePath,
            WebsiteBasePath: &websiteBasePath,
        },
        RecipeList: []supertokens.Recipe{
            passwordless.Init(plessmodels.TypeInput{
                FlowType: "MAGIC_LINK",
                ContactMethodEmail: plessmodels.ContactMethodEmailConfig{
                    Enabled: true,
                },
            }),
            session.Init(&sessmodels.TypeInput{
                GetTokenTransferMethod: func(req *http.Request, forCreateNewSession bool, userContext *map[string]interface{}) sessmodels.TokenTransferMethod {
                    return sessmodels.CookieTransferMethod
                },
            }),
        },
    })
    if err != nil {
        log.Printf("SuperTokens initialization failed: %v", err)
        return err
    }
    log.Println("SuperTokens initialized successfully")
    return nil
}


func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
        w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

func NewRouter() *mux.Router {
    router := mux.NewRouter()
    
    // Add CORS middleware
    router.Use(corsMiddleware)
    
    // Add logging middleware
    router.Use(func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            log.Printf("Incoming request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
            next.ServeHTTP(w, r)
            log.Printf("Completed request: %s %s", r.Method, r.URL.Path)
        })
    })
    
    router.HandleFunc("/", handleHome).Methods("GET", "OPTIONS")
    router.HandleFunc("/auth/login", handleLogin).Methods("POST", "OPTIONS")
    router.HandleFunc("/auth/verify", handleVerify).Methods("GET", "OPTIONS")
    router.HandleFunc("/auth/verify-code", handleVerifyCode).Methods("POST", "OPTIONS")
    return router
}

func handleHome(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    http.ServeFile(w, r, "static/index.html")
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
    log.Printf("Handling login request from %s", r.RemoteAddr)

    var input struct {
        Email string `json:"email"`
    }
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        log.Printf("Error decoding login request: %v", err)
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    log.Printf("Received login request for email: %s", input.Email)

    // Call SuperTokens API to generate the code
    result, err := passwordless.CreateCodeWithEmail("public", input.Email, nil)
    if err != nil {
        log.Printf("Error generating magic link: %v", err)
        http.Error(w, "Failed to generate magic link", http.StatusInternalServerError)
        return
    }

    if result.OK == nil {
        log.Printf("Error: No valid code returned by SuperTokens")
        http.Error(w, "Failed to generate magic link", http.StatusInternalServerError)
        return
    }

    preAuthSessionID := result.OK.PreAuthSessionID
    linkCode := result.OK.LinkCode

    // Construct the magic link
    host := r.Host
    magicLink := fmt.Sprintf(
        "http://%s/auth/verify?preAuthSessionId=%s&tenantId=public&linkCode=%s",
        host, preAuthSessionID, linkCode,
    )

    log.Printf("Generated magic link: %s", magicLink)

    // Respond with the magic link
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status":    "success",
        "link":      magicLink,
        "message":   "Magic link generated successfully",
        "tenantId":  "public",
        "sessionId": preAuthSessionID,
    })
}

func handleVerify(w http.ResponseWriter, r *http.Request) {
    preAuthSessionId := r.URL.Query().Get("preAuthSessionId")
    linkCode := r.URL.Query().Get("linkCode")

    if preAuthSessionId == "" || linkCode == "" {
        http.Error(w, "Missing required parameters", http.StatusBadRequest)
        return
    }

    // Call SuperTokens API to consume the code
    resp, err := passwordless.ConsumeCodeWithLinkCode(preAuthSessionId, linkCode, "public", nil)
    if err != nil {
        log.Printf("Error consuming code: %v", err)
        http.Error(w, "Invalid or expired link", http.StatusBadRequest)
        return
    }

    // Respond with success
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status":  "success",
        "message": "Login successful",
        "userId":  resp.OK.User.ID,
    })
}


func handleVerifyCode(w http.ResponseWriter, r *http.Request) {
    log.Printf("Handling verify-code request from %s", r.RemoteAddr)

    // Extract parameters from query string
    preAuthSessionId := r.URL.Query().Get("preAuthSessionId")
    linkCode := r.URL.Query().Get("linkCode")

    log.Printf("Attempting to verify code with preAuthSessionId: %s and linkCode: %s", preAuthSessionId, linkCode)

    if preAuthSessionId == "" || linkCode == "" {
        log.Printf("Missing required parameters")
        http.Error(w, "Missing preAuthSessionId or linkCode", http.StatusBadRequest)
        return
    }

    // Call SuperTokens API to verify the code
    resp, err := passwordless.ConsumeCodeWithLinkCode(preAuthSessionId, linkCode, "public", nil)
    if err != nil {
        log.Printf("Error consuming code: %v", err)
        http.Error(w, "Invalid or expired code", http.StatusBadRequest)
        return
    }

    log.Printf("Code verified successfully for user: %s", resp.OK.User.ID)

    // Create a new session
    sessionInfo, err := session.CreateNewSession(
        r,
        w,
        "public", // Fixed tenant ID
        resp.OK.User.ID,
        nil,
        nil,
    )
    if err != nil {
        log.Printf("Error creating session: %v", err)
        http.Error(w, "Failed to create session", http.StatusInternalServerError)
        return
    }

    log.Printf("Session created successfully for user: %s", sessionInfo.GetUserID())

    // Respond with success
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status":  "success",
        "message": "Logged in successfully",
        "userId":  sessionInfo.GetUserID(),
    })
}


