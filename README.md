# 🔮 Steganography Lab  

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)  
![JavaScript](https://img.shields.io/badge/JavaScript-F7DF1E?style=for-the-badge&logo=javascript&logoColor=black)  
An **advanced web-based steganography tool** that allows you to securely hide secret files within images, audio, video, and PDF files — protected by **AES-256 encryption**.  

---

## ✨ Live Demonstration  
🔗 [Try it here](http://ec2-3-106-122-161.ap-southeast-2.compute.amazonaws.com/)  

## 🎥 Demonstration Video  

Watch the demo on YouTube:  
🔗 [https://youtu.be/kE94xAlbDuE](https://youtu.be/kE94xAlbDuE)  

---

## 🚀 Features  

- 🔐 **Secure Encryption**  
  AES-256-GCM encryption with PBKDF2-derived keys ensures confidentiality & integrity.  

- 🖼️ **Multi-Format Carrier Support**  
  Hide data inside:  
  - Images (`.png`, `.jpg`)  
  - Audio (`.wav`, `.mp3`)  
  - Video (`.mp4`)  
  - Documents (`.pdf`)  

- 📂 **Flexible Payloads**  
  Supports:  
  - Text messages  
  - Images, audio, video, or PDF files  

- 💡 **Intuitive Web Interface**  
  Clean, modern, and user-friendly UI.  

- 🛡️ **Built-in Safeguards**  
  Prevents embedding excessively large payloads that could break stealth.  

---

## 🛠️ Technology Stack  

| Area      | Technologies |
|-----------|--------------|
| Backend   | Go (Golang), Gin Gonic |
| Frontend  | Vanilla JavaScript (ES6+), HTML5, CSS3 |
| Deployment| AWS EC2 (Frontend), Render (Backend API) |

---

## ⚙️ Getting Started  

### Prerequisites  
- Go **v1.18+**  
- Modern browser (Chrome, Firefox, etc.)  
- VS Code + Live Server extension  

---

### 1️⃣ Backend Setup (Go API)  

```bash
# Clone the repository
git clone https://github.com/kknguyen214/Prj_Seminar_Steganography.git
cd steganography-lab

# Install dependencies
go mod tidy

# Run the server
go run main.go

```
The API server will start and listen on **http://localhost:8080**.  
Keep this terminal running.  


### 2️⃣ Frontend Setup (Web Interface)  

#### Configure the API endpoint  
Open the **`script.js`** file in your code editor.  
Find the `API_BASE` constant and change it to point to your local server:  

```js
// Change this:
// const API_BASE = 'https://steganography-lab-backend.onrender.com/api';

// To this:
const API_BASE = 'http://localhost:8080/api';
```

### Launch the application  

- In VS Code, right-click on `index.html`  
- Select **"Open with Live Server"**  
- Your browser will open at: `http://127.0.0.1:5500`  

👉 You can now use the application locally!  

---

## 🔐 Security Highlights  

This project was built with **security as a top priority**:  

- **Passphrase-based Key Derivation (PBKDF2)**  
  - 10,000 iterations + random salt  
  - Derives a strong 256-bit encryption key from your passphrase  
  - Makes brute-force attacks computationally infeasible  

- **Authenticated Encryption (AES-GCM)**  
  - Ensures confidentiality (unreadable data)  
  - Guarantees integrity (tampering is detectable)  

- **Unique Salt per Encryption**  
  - A new random 16-byte salt is generated for each embedding  
  - Prevents rainbow table attacks  
  - Ensures same secret + same passphrase → different ciphertext  

---

## 🗺️ Roadmap  

Planned future enhancements:  

- **More Advanced Steganography**  
  Implement robust algorithms (e.g., F5 for JPEGs) resistant to steganalysis.  

- **Batch Processing**  
  Embed or extract data from multiple files at once.  

- **Client-Side Processing**  
  Use WebAssembly to perform encryption & steganography in-browser.  

- **Enhanced Media Support**  
  Improve compatibility with more codecs & formats.  

---

