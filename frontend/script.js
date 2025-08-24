// API Configuration
const API_BASE = 'http://localhost:8080/api'; // Change to your API URL

// Tab switching
function switchTab(tabName) {
    document.querySelectorAll('.tab').forEach(tab => tab.classList.remove('active'));
    document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));
    
    event.target.classList.add('active');
    document.getElementById(tabName + '-tab').classList.add('active');
}

// Toggle message input based on type
function toggleMessageInput() {
    const messageType = document.getElementById('message-type').value;
    
    // Hide all inputs
    document.getElementById('text-input').style.display = 'none';
    document.getElementById('image-input').style.display = 'none';
    document.getElementById('audio-input').style.display = 'none';
    
    // Show selected input
    if (messageType === 'text') {
        document.getElementById('text-input').style.display = 'block';
    } else if (messageType === 'image') {
        document.getElementById('image-input').style.display = 'block';
    } else if (messageType === 'audio') {
        document.getElementById('audio-input').style.display = 'block';
    }
}

// File input handlers
function setupFileInput(inputId, nameDisplayId) {
    const input = document.getElementById(inputId);
    const nameDisplay = document.getElementById(nameDisplayId);
    const label = input.nextElementSibling;
    
    input.addEventListener('change', function() {
        if (this.files && this.files[0]) {
            const fileName = this.files[0].name;
            const fileSize = (this.files[0].size / 1024 / 1024).toFixed(2) + ' MB';
            nameDisplay.innerHTML = `<br><strong>${fileName}</strong><br><small>${fileSize}</small>`;
            label.classList.add('has-file');
        } else {
            nameDisplay.innerHTML = '';
            label.classList.remove('has-file');
        }
    });
}

// Setup all file inputs
setupFileInput('host-file', 'host-file-name');
setupFileInput('secret-image', 'secret-image-name');
setupFileInput('secret-audio', 'secret-audio-name');
setupFileInput('extract-file', 'extract-file-name');

// Embed form handler
document.getElementById('embed-form').addEventListener('submit', async function(e) {
    e.preventDefault();
    
    const loading = document.getElementById('embed-loading');
    const result = document.getElementById('embed-result');
    
    loading.classList.add('show');
    result.innerHTML = '';
    
    try {
        const formData = new FormData();
        
        // Host file
        const hostFile = document.getElementById('host-file').files[0];
        formData.append('image', hostFile);
        
        // Message type and password
        const messageType = document.getElementById('message-type').value;
        const password = document.getElementById('embed-password').value;
        
        formData.append('message_type', messageType);
        formData.append('passphrase', password);
        
        // Add message content based on type
        if (messageType === 'text') {
            const text = document.getElementById('secret-text').value;
            formData.append('text', text);
        } else if (messageType === 'image') {
            const imageFile = document.getElementById('secret-image').files[0];
            if (!imageFile) {
                throw new Error('Please select an image file');
            }
            formData.append('message_image', imageFile);
        } else if (messageType === 'audio') {
            const audioFile = document.getElementById('secret-audio').files[0];
            if (!audioFile) {
                throw new Error('Please select an audio file');
            }
            formData.append('audio', audioFile);
        }
        
        const response = await fetch(`${API_BASE}/embed`, {
            method: 'POST',
            body: formData
        });
        
        if (response.ok) {
            const blob = await response.blob();
            const url = URL.createObjectURL(blob);
            
            // Get original filename and determine file extension
            const hostFile = document.getElementById('host-file').files[0];
            const originalName = hostFile.name;
            const fileExtension = originalName.split('.').pop().toLowerCase();
            const downloadFileName = `embedded_${Date.now()}.${fileExtension}`;
            
            result.innerHTML = `
                <div class="result-container">
                    <h4>✅ Success!</h4>
                    <p>Data has been successfully embedded in the file. The result file looks identical to the original but contains your hidden data.</p>
                    
                    <div class="form-group" style="margin-top: 20px;">
                        <label>🖼️ Result Preview:</label>
                        <img src="${url}" class="result-image" alt="Embedded File Result" style="max-width: 100%; height: auto; border: 2px solid #e9ecef; border-radius: 10px; display: block; margin: 10px auto;">
                    </div>
                    
                    <div class="form-group">
                        <label>📊 File Information:</label>
                        <div style="background: #f8f9fa; padding: 15px; border-radius: 8px; margin-top: 10px;">
                            <p><strong>Original File:</strong> ${originalName}</p>
                            <p><strong>File Size:</strong> ${(blob.size / 1024 / 1024).toFixed(2)} MB</p>
                            <p><strong>Status:</strong> <span style="color: #28a745; font-weight: bold;">✅ Data Successfully Hidden</span></p>
                        </div>
                    </div>
                    
                    <div style="text-align: center; margin-top: 25px; display: flex; gap: 15px; justify-content: center; flex-wrap: wrap;">
                        <a href="${url}" download="${downloadFileName}" class="btn btn-secondary">
                            💾 Download Result File
                        </a>
                    </div>
                </div>
            `;
        } else {
            const error = await response.json();
            throw new Error(error.message || 'Failed to embed data');
        }
        
    } catch (error) {
        result.innerHTML = `
            <div class="result-container error">
                <h4>❌ Error</h4>
                <p>${error.message}</p>
            </div>
        `;
    } finally {
        loading.classList.remove('show');
    }
});

// Extract form handler
document.getElementById('extract-form').addEventListener('submit', async function(e) {
    e.preventDefault();
    
    const loading = document.getElementById('extract-loading');
    const result = document.getElementById('extract-result');
    
    loading.classList.add('show');
    result.innerHTML = '';
    
    try {
        const formData = new FormData();
        
        const file = document.getElementById('extract-file').files[0];
        const password = document.getElementById('extract-password').value;
        
        formData.append('image', file);
        formData.append('passphrase', password);
        
        const response = await fetch(`${API_BASE}/extract`, {
            method: 'POST',
            body: formData
        });
        
        const data = await response.json();
        
        if (data.success) {
            const messageType = data.data.message_type;
            const content = data.data.content;
            
            let contentHtml = '';
            
            if (messageType === 'text') {
                contentHtml = `
                    <div class="form-group">
                        <label>📝 Extracted Text:</label>
                        <textarea class="form-control" rows="6" readonly>${content}</textarea>
                    </div>
                `;
            } else if (messageType === 'image') {
                contentHtml = `
                    <div class="form-group">
                        <label>🖼️ Extracted Image:</label>
                        <img src="data:image/png;base64,${content}" class="result-image" alt="Extracted Image">
                    </div>
                `;
            } else if (messageType === 'audio') {
                contentHtml = `
                    <div class="form-group">
                        <label>🎵 Extracted Audio:</label>
                        <audio controls class="result-audio">
                            <source src="data:audio/mpeg;base64,${content}" type="audio/mpeg">
                            Your browser does not support the audio element.
                        </audio>
                    </div>
                `;
            }
            
            result.innerHTML = `
                <div class="result-container">
                    <h4>✅ Extraction Successful!</h4>
                    <p><strong>Data Type:</strong> ${messageType}</p>
                    ${contentHtml}
                </div>
            `;
        } else {
            throw new Error(data.message || 'Failed to extract data');
        }
        
    } catch (error) {
        result.innerHTML = `
            <div class="result-container error">
                <h4>❌ Error</h4>
                <p>${error.message}</p>
            </div>
        `;
    } finally {
        loading.classList.remove('show');
    }
});