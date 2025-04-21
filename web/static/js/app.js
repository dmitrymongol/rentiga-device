// web/static/js/app.js
document.addEventListener('DOMContentLoaded', () => {
    // Получаем элементы при загрузке страницы
    const streamStatus = document.getElementById('streamStatus');
    const startBtn = document.getElementById('startBtn');
    const stopBtn = document.getElementById('stopBtn');
    const certForm = document.getElementById('certForm');
    const certInput = document.getElementById('certInput');
    const certButton = document.querySelector('#certForm button[type="submit"]');
    const connectionStatus = document.getElementById('connectionStatus');

    async function updateStatus() {
        try {
            const response = await fetch('/api/status');
            const status = await response.json();
            
            
            connectionStatus.className = `status-box ${status.connected ? 'connected' : 'disconnected'}`;
            connectionStatus.querySelector('.status-text').textContent = 
                `Connection: ${status.connected ? 'Connected' : 'Disconnected'}`;
    
            streamStatus.className = `status-box ${status.streaming ? 'streaming' : 'stopped'}`;
            streamStatus.querySelector('.status-text').textContent = 
                `Stream: ${status.streaming ? 'Active' : 'Inactive'}`;

            if (status.has_certificate) {
                certButton.textContent = 'Update Certificate';
            } else {
                certButton.textContent = 'Upload Certificate';
            }
        } catch (error) {
            console.error('Status update failed:', error);
            statusElement.textContent = 'Status: Error';
        }
    }

    // Обработчики событий
    startBtn.addEventListener('click', async () => {
        await fetch('/api/start', {method: 'POST'});
        await updateStatus();
    });

    stopBtn.addEventListener('click', async () => {
        await fetch('/api/stop', {method: 'POST'});
        await updateStatus();
    });

    certForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData();
        formData.append('certificate', certInput.files[0]);

        try {
            const response = await fetch('/api/upload-cert', {
                method: 'POST',
                body: formData
            });
            
            if (!response.ok) throw new Error('Upload failed');
            await updateStatus();
        } catch (error) {
            alert('Error uploading certificate: ' + error.message);
        }
    });

    // Первоначальное обновление статуса
    updateStatus();
    // Обновление статуса каждые 2 секунды
    setInterval(updateStatus, 2000);
});