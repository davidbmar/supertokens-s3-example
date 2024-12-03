package api

import (
    "encoding/json"
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
        log.Printf("Error decoding request body: %v", err)
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    log.Printf("Received login request for email: %s", input.Email)

    w.Header().Set("Content-Type", "application/json")
    
    link, err := passwordless.CreateMagicLinkByEmail("public", input.Email)
    if err != nil {
        log.Printf("Error creating magic link: %v", err)
        json.NewEncoder(w).Encode(map[string]string{
            "status": "error",
            "error":  err.Error(),
        })
        return
    }

    log.Printf("Successfully created magic link for %s: %s", input.Email, link)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "success",
        "link":   link,
    })
}


func handleVerify(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    w.Write([]byte(`
    <html>
        <head>
            <title>Verify Login</title>
        </head>
        <body>
            <h1>Verifying login...</h1>
            <div id="status"></div>
        </body>
        <script>
            document.addEventListener('DOMContentLoaded', function() {
                const linkCode = window.location.hash.substring(1);
                const preAuthSessionId = new URLSearchParams(window.location.search).get('preAuthSessionId');
                
                console.log("PreAuthSessionId:", preAuthSessionId);
                console.log("LinkCode:", linkCode);
                
                if (!preAuthSessionId || !linkCode) {
                    document.getElementById('status').innerHTML = '<h1>Error: Missing required parameters</h1>';
                    return;
                }

                console.log("Making verification request...");
                fetch('/auth/verify-code', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        preAuthSessionId: preAuthSessionId,
                        linkCode: linkCode,
                        tenantId: "public"
                    })
                })
                .then(response => {
                    console.log("Response status:", response.status);
                    return response.text().then(text => {
                        try {
                            return JSON.parse(text);
                        } catch (e) {
                            throw new Error('Server response: ' + text);
                        }
                    });
                })
                .then(data => {
                    console.log("Verification response:", data);
                    if (data.status === 'success') {
                        document.getElementById('status').innerHTML = '<h1>Successfully logged in!</h1><p>User ID: ' + data.userId + '</p>';
                    } else {
                        document.getElementById('status').innerHTML = '<h1>Error: ' + (data.message || 'Unknown error') + '</h1>';
                    }
                })
                .catch(error => {
                    console.error('Verification error:', error);
                    document.getElementById('status').innerHTML = '<h1>Error verifying link</h1><p>' + error.message + '</p>';
                });
            });
        </script>
    </html>
    `))
}
func handleVerifyCode(w http.ResponseWriter, r *http.Request) {
    log.Printf("Handling verify-code request from %s", r.RemoteAddr)
    
    var input struct {
        PreAuthSessionId string `json:"preAuthSessionId"`
        LinkCode        string `json:"linkCode"`
    }

    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        log.Printf("Error decoding verify request body: %v", err)
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    log.Printf("Attempting to verify code with preAuthSessionId: %s and linkCode: %s", input.PreAuthSessionId, input.LinkCode)

    if input.PreAuthSessionId == "" || input.LinkCode == "" {
        log.Printf("Missing required parameters")
        http.Error(w, "Missing preAuthSessionId or linkCode", http.StatusBadRequest)
        return
    }

    resp, err := passwordless.ConsumeCodeWithLinkCode(
        input.PreAuthSessionId,
        input.LinkCode,
        "public",  // Fixed tenant ID
        nil,
    )
    if err != nil {
        log.Printf("Error consuming code: %v", err)
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    log.Printf("Code verified successfully, creating session for user: %s", resp.OK.User.ID)

    sessionInfo, err := session.CreateNewSession(
        r,
        w,
        "public",  // Fixed tenant ID
        resp.OK.User.ID,
        nil,
        nil,
    )
    if err != nil {
        log.Printf("Error creating session: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    log.Printf("Session created successfully for user: %s", sessionInfo.GetUserID())

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status":  "success",
        "message": "Logged in successfully",
        "userId":  sessionInfo.GetUserID(),
    })
}

