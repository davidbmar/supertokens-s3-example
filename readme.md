Here’s a comprehensive documentation of everything we’ve done, including setup steps, commands, and explanations for running the SuperTokens-based authentication system:

---

# **SuperTokens Authentication System Setup**

This guide explains how to set up a SuperTokens-based authentication system on an AWS EC2 instance with a PostgreSQL database and integrate it with a Go backend.

---

## **Prerequisites**
1. **AWS EC2 Instance**:
   - An Ubuntu 20.04 (or similar) instance with public IP.
   - Security Group configured to allow:
     - Port `22` (for SSH)
     - Port `8080` (for backend API)
     - Port `3567` (optional for direct access to SuperTokens, internal use recommended).
2. **Docker** installed on the EC2 instance:
   ```bash
   sudo apt-get update
   sudo apt-get install -y docker.io
   sudo systemctl enable docker
   sudo systemctl start docker
   ```
3. **Go** installed on your development machine and EC2:
   ```bash
   sudo apt-get install -y golang
   ```

---

## **Step 1: Setting Up Docker Containers**

### **Setup Script**
The provided `sample_docker_container_setup.sh` automates the deployment of PostgreSQL and SuperTokens containers. Place the script on your EC2 instance and execute it.

### **Steps to Run the Script**
1. **Upload the script** to your EC2 instance:
   ```bash
   scp -i "your-key.pem" sample_docker_container_setup.sh ubuntu@<your-ec2-ip>:/home/ubuntu/
   ```
2. **Run the script**:
   ```bash
   chmod +x sample_docker_container_setup.sh
   ./sample_docker_container_setup.sh
   ```
3. **What the Script Does**:
   - Creates a Docker network named `supertokens`.
   - Deploys a PostgreSQL container:
     - Username: `supertokens`
     - Password: `your-strong-password`
     - Database: `supertokens`
   - Deploys a SuperTokens container:
     - Exposes port `3567`.
     - Connects to PostgreSQL for data storage.

4. **Verify Running Containers**:
   ```bash
   docker ps
   ```

---

## **Step 2: Setting Up the Go Backend**

The backend is responsible for handling authentication requests from clients and forwarding them to SuperTokens.

### **File Structure**
Your project directory should look like this:
```
transcription-service/
├── cmd/
│   └── server/
│       └── main.go         # Backend entry point
├── internal/
│   ├── auth/              # Authentication logic
│   │   └── auth.go
│   ├── api/               # HTTP handlers
│   │   └── handlers.go
│   ├── storage/           # S3 operations (optional)
│   │   └── storage.go
├── config/
│   └── config.go          # Configuration management
├── scripts/
│   └── sample_docker_container_setup.sh # Docker setup script
├── go.mod                 # Go module dependencies
└── static/
    └── index.html         # Frontend for testing
```

### **Running the Backend**
1. SSH into the EC2 instance.
2. Navigate to the backend directory:
   ```bash
   cd /home/ubuntu/go/src/transcription-service
   ```
3. Start the backend server:
   ```bash
   go run cmd/server/main.go
   ```
4. The server will start listening on port `8080`.

---

## **Step 3: Testing Authentication**

### **Login API**
1. Send a login request from your local machine:
   ```bash
   curl -X POST http://<your-ec2-public-ip>:8080/auth/login \
   -H "Content-Type: application/json" \
   -d '{"email": "user@example.com"}'
   ```
2. Expected Response:
   ```json
   {
       "link": "http://<your-ec2-public-ip>:8080/auth/verify?preAuthSessionId=<id>&tenantId=public#<linkCode>",
       "status": "success"
   }
   ```

### **Verify API**
1. Open the magic link from the response in a browser or manually call the verify endpoint:
   ```bash
   curl -X GET "http://<your-ec2-public-ip>:8080/auth/verify?preAuthSessionId=<id>&tenantId=public#<linkCode>"
   ```
2. The backend should validate the session via SuperTokens.

---

## **Step 4: Frontend Integration**

The `index.html` file in your `static/` directory provides a basic frontend for testing magic link authentication. Update the `API_URL` in the file to your EC2 public IP:
```javascript
const API_URL = 'http://<your-ec2-public-ip>:8080';
```

Deploy the file using any web server (e.g., Nginx) or test it locally.

---

## **Security Considerations**
1. **SuperTokens Accessibility**:
   - Restrict port `3567` to be accessible only within the EC2 instance.
   - Use a private VPC network or security group rules to ensure the backend communicates internally with SuperTokens.
2. **HTTPS**:
   - Set up SSL for the backend using tools like Let's Encrypt.
   - For testing, you can use an Nginx reverse proxy to enable HTTPS.
3. **Environment Variables**:
   - Store sensitive data (e.g., database credentials, API keys) in environment variables or AWS Secrets Manager.

---

## **Common Commands**

### **Docker Management**
- **Check running containers**:
  ```bash
  docker ps
  ```
- **View logs**:
  ```bash
  docker logs supertokens-core
  docker logs supertokens-postgres
  ```

### **Backend Testing**
- **Send login request**:
  ```bash
  curl -X POST http://<your-ec2-public-ip>:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com"}'
  ```

---

## **Next Steps**
- Integrate S3 for file storage (if needed).
- Add user session handling in the backend for authenticated operations.
- Deploy the backend and frontend using a CI/CD pipeline for production readiness.

