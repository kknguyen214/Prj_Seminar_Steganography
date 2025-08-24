// API Configuration
const API_BASE = 'http://localhost:8080/api'; // Thay đổi URL API của bạn

// Tab switching
function switchTab(tabName) {
    document.querySelectorAll('.tab').forEach(tab => tab.classList.remove('active'));
    document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));
    
    event.target.classList.add('active');
    document.getElementById(tabName + '-tab').classList.add('active');
}

// Toggle media input for embed
function toggleMediaInput() {
    const mediaType = document.getElementById('media-type').value;
    
    // Hide all inputs
    document.getElementById('image-host-input').style.display = 'none';
    document.getElementById('video-host-input').style.display = 'none';
    document.getElementById('audio-host-input').style.display = 'none';
    
    // Show selected input
    if (mediaType === 'image') {
        document.getElementById('image-host-input').style.display = 'block';
    } else if (mediaType === 'video') {
        document.getElementById('video-host-input').style.display = 'block';
    } else if (mediaType === 'audio') {
        document.getElementById('audio-host-input').style.display = 'block';
    }
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

// Toggle extract input based on type
function toggleExtractInput() {
    const mediaType = document.getElementById('extract-media-type').value;
    
    // Hide all inputs
    document.getElementById('extract-image-input').style.display = 'none';
    document.getElementById('extract-video-input').style.display = 'none';
    document.getElementById('extract-audio-input').style.display = 'none';
    
    // Show selected input
    if (mediaType === 'image') {
        document.getElementById('extract-image-input').style.display = 'block';
    } else if (mediaType === 'video') {
        document.getElementById('extract-video-input').style.display = 'block';
    } else if (mediaType === 'audio') {
        document.getElementById('extract-audio-input').style.display = 'block';
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

// Setup all file inputs when page loads
document.addEventListener('DOMContentLoaded', function() {
    setupFileInput('host-image', 'host-image-name');
    setupFileInput('host-video', 'host-video-name');
    setupFileInput('host-audio', 'host-audio-name');
    setupFileInput('secret-image', 'secret-image-name');
    setupFileInput('secret-audio', 'secret-audio-name');
    setupFileInput('extract-image', 'extract-image-name');
    setupFileInput('extract-video', 'extract-video-name');
    setupFileInput('extract-audio', 'extract-audio-name');
});

// Get appropriate file field name for API
function getFileFieldName(mediaType) {
    switch(mediaType) {
        case 'image': return 'image';
        case 'video': return 'video';
        case 'audio': return 'audio_carrier';
        default: return 'image';
    }
}

// Get appropriate host file input
function getHostFileInput(mediaType) {
    switch(mediaType) {
        case 'image': return document.getElementById('host-image');
        case 'video': return document.getElementById('host-video');
        case 'audio': return document.getElementById('host-audio');
        default: return null;
    }
}

// Get appropriate extract file input
function getExtractFileInput(mediaType) {
    switch(mediaType) {
        case 'image': return document.getElementById('extract-image');
        case 'video': return document.getElementById('extract-video');
        case 'audio': return document.getElementById('extract-audio');
        default: return null;
    }
}

// Create download link for extracted files
function createDownloadLink(content, messageType, filename = null) {
    if (messageType === 'text') {
        return `<textarea class="form-control" rows="6" readonly>${content}</textarea>`;
    }
    
    // For binary data (image/audio), create download link
    const defaultFilename = filename || (messageType === 'image' ? 'extracted_image.png' : 'extracted_audio.wav');
    const mimeType = messageType === 'image' ? 'image/png' : 'audio/wav';
    
    try {
        const binaryData = atob(content);
        const bytes = new Uint8Array(binaryData.length);
        for (let i = 0; i < binaryData.length; i++) {
            bytes[i] = binaryData.charCodeAt(i);
        }
        const blob = new Blob([bytes], { type: mimeType });
        const url = URL.createObjectURL(blob);
        
        let mediaElement = '';
        if (messageType === 'image') {
            mediaElement = `<img src="${url}" class="result-image" alt="Extracted Image">`;
        } else {
            mediaElement = `<audio controls class="result-audio"><source src="${url}" type="${mimeType}"></audio>`;
        }
        
        return `
            ${mediaElement}
            <div style="text-align: center; margin-top: 15px;">
                <a href="${url}" download="${defaultFilename}" class="download-btn">
                    💾 Download ${messageType === 'image' ? 'Image' : 'Audio'}
                </a>
            </div>
        `;
    } catch (error) {
        return `<p class="error">Error processing ${messageType} data</p>`;
    }
}

// Embed form handler
document.addEventListener('DOMContentLoaded', function() {
    document.getElementById('embed-form').addEventListener('submit', async function(e) {
        e.preventDefault();
        
        const loading = document.getElementById('embed-loading');
        const result = document.getElementById('embed-result');
        
        loading.classList.add('show');
        result.innerHTML = '';
        
        try {
            const formData = new FormData();
            
            // Get media type and host file
            const mediaType = document.getElementById('media-type').value;
            const hostFileInput = getHostFileInput(mediaType);
            const hostFile = hostFileInput.files[0];
            
            if (!hostFile) {
                throw new Error('Please select a host file');
            }
            
            // Add host file with correct field name
            const fileFieldName = getFileFieldName(mediaType);
            formData.append(fileFieldName, hostFile);
            
            // Add media type and message type
            formData.append('media_type', mediaType);
            
            const messageType = document.getElementById('message-type').value;
            const password = document.getElementById('embed-password').value;
            
            formData.append('message_type', messageType);
            formData.append('passphrase', password);
            
            // Add message content based on type
            if (messageType === 'text') {
                const text = document.getElementById('secret-text').value;
                if (!text.trim()) {
                    throw new Error('Please enter secret text');
                }
                formData.append('text', text);
            } else if (messageType === 'image') {
                const imageFile = document.getElementById('secret-image').files[0];
                if (!imageFile) {
                    throw new Error('Please select a secret image file');
                }
                formData.append('message_image', imageFile);
            } else if (messageType === 'audio') {
                const audioFile = document.getElementById('secret-audio').files[0];
                if (!audioFile) {
                    throw new Error('Please select a secret audio file');
                }
                formData.append('message_audio', audioFile);
            }
            
            const response = await fetch(`${API_BASE}/embed`, {
                method: 'POST',
                body: formData
            });
            
            if (response.ok) {
                const blob = await response.blob();
                const url = URL.createObjectURL(blob);
                
                // Get original filename and create download name
                const originalName = hostFile.name;
                const downloadName = `embedded_${originalName}`;
                
                result.innerHTML = `
                    <div class="result-container">
                        <h4>✅ Success!</h4>
                        <p>Data has been successfully embedded in the ${mediaType} file.</p>
                        <div class="download-info">
                            <strong>📁 File Info:</strong><br>
                            Original: ${originalName}<br>
                            Size: ${(blob.size / 1024 / 1024).toFixed(2)} MB<br>
                            Type: ${mediaType.toUpperCase()}<span class="media-type-indicator">${mediaType}</span>
                        </div>
                        <div style="text-align: center; margin-top: 20px;">
                            <a href="${url}" download="${downloadName}" class="btn btn-secondary">
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
            
            const mediaType = document.getElementById('extract-media-type').value;
            const fileInput = getExtractFileInput(mediaType);
            const file = fileInput.files[0];
            const password = document.getElementById('extract-password').value;
            
            if (!file) {
                throw new Error('Please select a file');
            }
            
            // Use correct field name based on media type
            const fieldName = getFileFieldName(mediaType);
            formData.append(fieldName, file);
            formData.append('media_type', mediaType);
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
                            ${createDownloadLink(content, messageType)}
                        </div>
                    `;
                } else if (messageType === 'image') {
                    contentHtml = `
                        <div class="form-group">
                            <label>🖼️ Extracted Image:</label>
                            ${createDownloadLink(content, messageType)}
                        </div>
                    `;
                } else if (messageType === 'audio') {
                    contentHtml = `
                        <div class="form-group">
                            <label>🎵 Extracted Audio:</label>
                            ${createDownloadLink(content, messageType)}
                        </div>
                    `;
                }
                
                result.innerHTML = `
                    <div class="result-container">
                        <h4>✅ Extraction Successful!</h4>
                        <p><strong>Data Type:</strong> ${messageType.toUpperCase()}</p>
                        <p><strong>Source:</strong> ${mediaType.toUpperCase()} file</p>
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
});