# API Key Limiter Management
Tool used to create and manage configurations for the [API Key Limiter proxy](https://github.com/IgorPidik/api-key-limiter). Users can create multiple projects and configurations. A number of headers and their values can be specified for each configuration. A unique proxy URL will be generated. Proxied requests will be processed, the header values will be added or updated and the modified request will be forwarded to the original host.


<img width="1512" alt="Screenshot 2025-02-01 at 21 42 41" src="https://github.com/user-attachments/assets/5880bd9d-3e5e-4db7-9987-0059f24eec7d" />
<img width="1512" alt="Screenshot 2025-02-01 at 21 43 25" src="https://github.com/user-attachments/assets/6f18fe2d-0165-4f91-a5a3-c2daea5b1122" />


## Setup
### 1. Setup .env file
Create `.env` file and update placeholder values
```bash
$ cp .env.example .env
```

`SECRET_KEY` is used for data encryption, you can come up with your own or you can generate one with:
```bash
$ make generate-secret-key
```
Please note that the secret key here and the one in your [proxy](https://github.com/IgorPidik/api-key-limiter) `.env` file must match.

### 2. Migrate the DB
```bash
$ make migrate
```

### 3. Run the project
```bash
$ make build
$ ./main
```

Navigate to [http://localhost:8080/projects](http://localhost:8080/projects)
