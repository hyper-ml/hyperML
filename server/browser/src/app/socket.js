import storage from 'local-storage-fallback'


const makeSocket = (ws, callback) => {
    if (!ws || ws.readyState === WebSocket.CLOSED) {
        ws = new WebSocket("ws://localhost:8888/user/websocket");
 
        ws.onopen = () => {
            console.log('websocket connected');
            
            var auth_msg = {
                type: "authenticate",
                token: storage.getItem('token')  
            }

            ws.send(JSON.stringify(auth_msg))
        }
    
        ws.onclose = e => {
            console.log('websocket closed due to ', e)
        }
    
        ws.onerror = err => {
            console.error('websocket failed: ', err.message)
            ws.close()
        }
        
    }
    return ws
}

export default makeSocket;