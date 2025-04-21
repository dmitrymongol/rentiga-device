async function updateStatus() {
    const response = await fetch('/api/status');
    const status = await response.json();
    document.getElementById('status').textContent = `Status: ${status.streaming ? 'Streaming' : 'Stopped'}`;
}

document.getElementById('startBtn').addEventListener('click', async () => {
    await fetch('/api/start', {method: 'POST'});
    await updateStatus();
});

document.getElementById('stopBtn').addEventListener('click', async () => {
    await fetch('/api/stop', {method: 'POST'});
    await updateStatus();
});

// Обновление статуса каждые 2 секунды
setInterval(updateStatus, 2000);
updateStatus();