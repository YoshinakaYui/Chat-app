export function createWebSocket(onMessage: (msg: any) => void) {
    const ws = new WebSocket("ws://localhost:8080/ws");
  
    ws.onopen = () => {
      console.log("WebSocket接続成功");
    };
  
    ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      onMessage(message);
    };
  
    ws.onerror = (error) => {
      console.error("WebSocketエラー:", error);
    };
  
    ws.onclose = () => {
      console.log("WebSocket接続が切断されました");
    };
  
    return ws;
  }
  