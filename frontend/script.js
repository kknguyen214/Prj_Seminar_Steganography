// API Configuration
const API_BASE = 'http://localhost:8080/api';

// Tab switching
function switchTab(tabName) {
    document.querySelectorAll('.tab').forEach(tab => tab.classList.remove('active'));
    document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));
    
    event.target.classList.add('active');
    document.getElementById(tabName + '-tab').classList.add('active');
}

// Unified function to update all file inputs based on current selections
function updateFileInputs() {
    updateHostFileInput();
    updateMessageInputs();
}

// Update host file input based on selected media type
function updateHostFileInput() {
    const mediaType = document.getElementById('host-media-type').value;
    const fileInput = document.getElementById('host-file');
    const fileLabel = document.getElementById('host-file-label');
    const fileName = document.getElementById('host-file-name');
    const fileGroup = document.getElementById('host-file-group');
    
    // Clear previous selection
    fileInput.value = '';
    fileName.innerHTML = '';
    fileLabel.classList.remove('has-file');
    
    const fileTypeConfig = {
        'image': {
            accept: 'image/*',
            label: '🖼️ Select Image Container',
            display: true
        },
        'audio': {
            accept: 'audio/*',
            label: '🎵 Select Audio Container',
            display: true
        },
        'video': {
            accept: 'video/*',
            label: '🎬 Select Video Container',
            display: true
        }
    };

    if (fileTypeConfig[mediaType]) {
        const config = fileTypeConfig[mediaType];
        fileGroup.style.display = 'block';
        fileInput.accept = config.accept;
        fileInput.disabled = false;
        fileLabel.innerHTML = `${config.label}<div id="host-file-name"></div>`;
    } else {
        fileGroup.style.display = 'none';
        fileInput.accept = '';
        fileInput.disabled = true;
    }
}

// Update message inputs based on selected message type
function updateMessageInputs() {
    const messageType = document.getElementById('message-type').value;
    
    // Hide all inputs
    const inputGroups = ['text-input', 'image-input', 'audio-input', 'video-input'];
    inputGroups.forEach(groupId => {
        document.getElementById(groupId).style.display = 'none';
    });
    
    // Show selected input
    if (messageType && inputGroups.includes(messageType + '-input')) {
        document.getElementById(messageType + '-input').style.display = 'block';
    }
}

// Enhanced file input handler with better file info display
function setupFileInput(inputId, nameDisplayId) {
    const input = document.getElementById(inputId);
    const nameDisplay = document.getElementById(nameDisplayId);
    const label = input.nextElementSibling;
    
    input.addEventListener('change', function() {
        if (this.files && this.files[0]) {
            const file = this.files[0];
            const fileName = file.name;
            const fileSize = (file.size / 1024 / 1024).toFixed(2);
            const fileType = file.type || 'Unknown';
            
            nameDisplay.innerHTML = `
                <br><strong style="color: #60a5fa;">${fileName}</strong>
                <br><small style="color: #94a3b8;">${fileSize} MB • ${fileType}</small>
            `;
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
setupFileInput('secret-video', 'secret-video-name');
setupFileInput('extract-image-file', 'extract-image-file-name');
setupFileInput('extract-audio-file', 'extract-audio-file-name');
setupFileInput('extract-video-file', 'extract-video-file-name');


// Enhanced embed form handler
document.getElementById('embed-form').addEventListener('submit', async function(e) {
    console.log("NHAN INIT");
    e.preventDefault();
    
    const loading = document.getElementById('embed-loading');
    const result = document.getElementById('embed-result');
    
    loading.classList.add('show');
    result.innerHTML = '';
    
    try {
        const formData = new FormData();
        
        // Host file validation and setup
        const mediaType = document.getElementById('host-media-type').value;
        if (!mediaType) throw new Error('Please select a media type');
        
        formData.append('media_type', mediaType);
        const hostFile = document.getElementById('host-file').files[0];
        if (!hostFile) throw new Error('Please select a host file');
        
        // Add host file with appropriate field name
        const hostFileFieldMap = {
            'image': 'carrier_image',
            'audio': 'carrier_audio',
            'video': 'carrier_video'
        };
        formData.append(hostFileFieldMap[mediaType], hostFile);
        
        // Message type and password validation
        const messageType = document.getElementById('message-type').value;
        const password = document.getElementById('embed-password').value;
        
        if (!messageType) throw new Error('Please select a payload type');
        if (!password) throw new Error('Please enter a password');

        formData.append('message_type', messageType);
        formData.append('passphrase', password);
        
        // Add message content based on type
        if (messageType === 'text') {
            const text = document.getElementById('secret-text').value;
            if (!text.trim()) throw new Error('Please enter secret text');
            formData.append('text', text);
        } else if (messageType === 'image') {
            const imageFile = document.getElementById('secret-image').files[0];
            if (!imageFile) throw new Error('Please select an image file');
            formData.append('message_image', imageFile);
        } else if (messageType === 'audio') {
            const audioFile = document.getElementById('secret-audio').files[0];
            if (!audioFile) throw new Error('Please select an audio file');
            formData.append('message_audio', audioFile);
        } else if (messageType === 'video') {
            const videoFile = document.getElementById('secret-video').files[0];
            if (!videoFile) throw new Error('Please select a video file');
            formData.append('message_video', videoFile);
        }

        const response = await fetch(`${API_BASE}/embed`, {
            method: 'POST',
            body: formData
        });
        
        if (response.ok) {
            const blob = await response.blob();
            const url = URL.createObjectURL(blob);
            
            // Enhanced result display
            const originalName = hostFile.name;
            const fileExtension = originalName.split('.').pop().toLowerCase();
            const downloadFileName = `steganography_${Date.now()}.${fileExtension}`;
            
            result.innerHTML = `
                <div class="result-container">
                    <h4>✅ Embedding Successful</h4>
                    <p style="color: #cbd5e1; margin-bottom: 20px;">Data has been successfully embedded using advanced steganographic algorithms. The container appears identical to the original while securely concealing your payload.</p>
                    
                    <div class="form-group">
                        <label style="color: #60a5fa;">🖼️ Steganographic Container:</label>
                        <img src="${url}" class="result-image" alt="Steganographic Container" loading="lazy">
                    </div>
                    
                    <div class="file-info">
                        <p><strong>Original File:</strong> ${originalName}</p>
                        <p><strong>Container Size:</strong> ${(blob.size / 1024 / 1024).toFixed(2)} MB</p>
                        <p><strong>Payload Type:</strong> ${messageType.toUpperCase()}</p>
                        <p><strong>Status:</strong> <span style="color: #34d399; font-weight: bold;">🔐 Data Successfully Concealed</span></p>
                        <p><strong>Security:</strong> <span style="color: #a78bfa;">AES Encrypted</span></p>
                    </div>
                    
                    <div class="download-section">
                        <a href="${url}" download="${downloadFileName}" class="btn btn-secondary">
                            💾 Download Container
                        </a>
                    </div>
                </div>
            `;
        } else {
            const error = await response.json();
            throw new Error(error.message || 'Embedding process failed');
        }
        
    } catch (error) {
        result.innerHTML = `
            <div class="result-container error">
                <h4>❌ Embedding Failed</h4>
                <p style="color: #f87171;">${error.message}</p>
                <div style="margin-top: 16px; padding: 16px; background: rgba(239, 68, 68, 0.1); border-radius: 8px; border-left: 4px solid #ef4444;">
                    <p style="color: #fca5a5; margin: 0;"><strong>Troubleshooting:</strong></p>
                    <ul style="color: #fca5a5; margin: 8px 0 0 20px;">
                        <li>Ensure all required fields are filled</li>
                        <li>Check file format compatibility</li>
                        <li>Verify file size limitations</li>
                        <li>Confirm network connectivity</li>
                    </ul>
                </div>
            </div>
        `;
    } finally {
        loading.classList.remove('show');
    }
});

// Enhanced extract form handler
document.getElementById('extract-form').addEventListener('submit', async function(e) {
    e.preventDefault();
    
    const loading = document.getElementById('extract-loading');
    const result = document.getElementById('extract-result');
    
    loading.classList.add('show');
    result.innerHTML = '';
    
    try {
        const formData = new FormData();

        const mediaType = document.getElementById('extract-media-type').value;
        if (!mediaType) throw new Error('Please select a media type');
        
        formData.append('media_type', mediaType);
        
        let file = null;

        if (mediaType === 'image') {
            file = document.getElementById('extract-image-file').files[0];
        } else if (mediaType === 'audio') {
            file = document.getElementById('extract-audio-file').files[0];
        } else if (mediaType === 'video') {
            file = document.getElementById('extract-video-file').files[0];
        }

        if (!file) throw new Error('Please select a container file');

        const password = document.getElementById('extract-password').value;
        
        if (!file) throw new Error('Please select a container file');
        if (!password) throw new Error('Please enter the decryption password');
        
        const hostFileFieldMap = {
            'image': 'image',
            'audio': 'audio',
            'video': 'video'
        };
        formData.append(hostFileFieldMap[mediaType], file);
        
        formData.append('passphrase', password);
        
        const response = await fetch(`${API_BASE}/extract`, {
            method: 'POST',
            body: formData
        });
        
        const data = await response.json();
        console.log(data);
        
        if (data.success) {
            const messageType = data.message_type;
            const content = data.content;

            let contentHtml = '';
            let downloadButton = '';
            
            if (messageType === 'text') {
                // console.log("JSHDJHJDSH");
                contentHtml = `
                    <div class="form-group">
                        <label style="color: #60a5fa;">📝 Extracted Text Payload:</label>
                        <textarea class="form-control" rows="6" readonly style="background: rgba(30, 41, 59, 0.8); color: #e2e8f0;">${content}</textarea>
                    </div>
                `;
                
                // console.log("OKOK");
                // Create text download
                const textBlob = new Blob([content], { type: 'text/plain' });
                const textUrl = URL.createObjectURL(textBlob);
                // console.log("OKADJHJJDOK");
                downloadButton = `
                    <a href="${textUrl}" download="extracted_text_${Date.now()}.txt" class="btn btn-secondary">
                        💾 Download Text
                    </a>
                `;
                console.log("TEXT");
            } else if (messageType === 'image') {
                const imageUrl = `data:image/png;base64,${content}`;
                contentHtml = `
                    <div class="form-group">
                        <label style="color: #60a5fa;">🖼️ Extracted Image Payload:</label>
                        <img src="${imageUrl}" class="result-image" alt="Extracted Image" loading="lazy">
                    </div>
                `;
                downloadButton = `
                    <a href="${imageUrl}" download="extracted_image_${Date.now()}.png" class="btn btn-secondary">
                        💾 Download Image
                    </a>
                `;
            } else if (messageType === 'audio') {
                const audioUrl = `data:audio/mpeg;base64,${content}`;
                contentHtml = `
                    <div class="form-group">
                        <label style="color: #60a5fa;">🎵 Extracted Audio Payload:</label>
                        <audio controls class="result-audio" style="width: 100%; margin-top: 12px;">
                            <source src="${audioUrl}" type="audio/mpeg">
                            <source src="${audioUrl}" type="audio/wav">
                            <source src="${audioUrl}" type="audio/ogg">
                            Your browser does not support the audio element.
                        </audio>
                    </div>
                `;
                downloadButton = `
                    <a href="${audioUrl}" download="extracted_audio_${Date.now()}.mp3" class="btn btn-secondary">
                        💾 Download Audio
                    </a>
                `;
            } else if (messageType === 'video') {
                const videoUrl = `data:video/mp4;base64,${content}`;
                contentHtml = `
                    <div class="form-group">
                        <label style="color: #60a5fa;">🎬 Extracted Video Payload:</label>
                        <video controls class="result-video" style="width: 100%; margin-top: 12px;">
                            <source src="${videoUrl}" type="video/mp4">
                            <source src="${videoUrl}" type="video/webm">
                            <source src="${videoUrl}" type="video/ogg">
                            Your browser does not support the video element.
                        </video>
                    </div>
                `;
                downloadButton = `
                    <a href="${videoUrl}" download="extracted_video_${Date.now()}.mp4" class="btn btn-secondary">
                        💾 Download Video
                    </a>
                `;
            }
            
            result.innerHTML = `
                <div class="result-container">
                    <h4>🔓 Extraction Successful</h4>
                    <p style="color: #cbd5e1; margin-bottom: 20px;">Steganographic analysis complete. Hidden payload successfully recovered and decrypted.</p>
                    
                    <div class="file-info">
                        <p><strong>Container:</strong> ${file.name}</p>
                        <p><strong>Payload Type:</strong> ${messageType.toUpperCase()}</p>
                        <p><strong>Extraction Status:</strong> <span style="color: #34d399; font-weight: bold;">✅ Success</span></p>
                        <p><strong>Security:</strong> <span style="color: #a78bfa;">AES Decrypted</span></p>
                    </div>
                    
                    ${contentHtml}
                    
                    <div class="download-section">
                        ${downloadButton}
                    </div>
                </div>
            `;
        } else {
            throw new Error(data.message || 'Extraction process failed');
        }
        
    } catch (error) {
        result.innerHTML = `
            <div class="result-container error">
                <h4>🔒 Extraction Failed</h4>
                <p style="color: #f87171;">${error.message}</p>
                <div style="margin-top: 16px; padding: 16px; background: rgba(239, 68, 68, 0.1); border-radius: 8px; border-left: 4px solid #ef4444;">
                    <p style="color: #fca5a5; margin: 0;"><strong>Common Issues:</strong></p>
                    <ul style="color: #fca5a5; margin: 8px 0 0 20px;">
                        <li>Incorrect decryption password</li>
                        <li>File does not contain embedded data</li>
                        <li>Container file may be corrupted</li>
                        <li>Incompatible steganographic format</li>
                    </ul>
                </div>
            </div>
        `;
    } finally {
        loading.classList.remove('show');
    }
});

// Initialize the application
document.addEventListener('DOMContentLoaded', function() {
    // Add smooth loading animation
    document.body.style.opacity = '0';
    setTimeout(() => {
        document.body.style.transition = 'opacity 0.5s ease-in-out';
        document.body.style.opacity = '1';
    }, 100);
});

// Add keyboard navigation support
document.addEventListener('keydown', function(e) {
    if (e.key === 'Tab') {
        // Enhanced tab navigation for better accessibility
        const focusableElements = document.querySelectorAll(
            'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
        );
        const firstElement = focusableElements[0];
        const lastElement = focusableElements[focusableElements.length - 1];
        
        if (e.shiftKey && document.activeElement === firstElement) {
            lastElement.focus();
            e.preventDefault();
        } else if (!e.shiftKey && document.activeElement === lastElement) {
            firstElement.focus();
            e.preventDefault();
        }
    }
});

function updateExtractInputs() {
    const mediaType = document.getElementById('extract-media-type').value;
    const inputGroups = ['extract-image-input', 'extract-audio-input', 'extract-video-input'];

    inputGroups.forEach(groupId => {
        const el = document.getElementById(groupId);
        if (el) el.style.display = 'none';
    });

    const selectedEl = document.getElementById(`extract-${mediaType}-input`);
    if (selectedEl) selectedEl.style.display = 'block';
}


window.switchTab = switchTab;
window.updateFileInputs = updateFileInputs;
window.updateHostFileInput = updateHostFileInput;
window.updateMessageInputs = updateMessageInputs;
window.updateExtractInputs = updateExtractInputs;