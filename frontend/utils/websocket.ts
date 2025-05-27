let listeners: ((msg: any) => void)[] = [];
let socket: WebSocket;


//export function connectWebSocket(onMessage: (msg: any) => void) {
export function connectWebSocket() {
    if (!socket || socket.readyState !== WebSocket.OPEN){
    socket = new WebSocket("ws://localhost:8080/ws");

    socket.onopen = () => {
      console.log("WebSocket接続成功");
    };
  
    socket.onmessage = (event) => {
      const message = JSON.parse(event.data);
      console.log("listeners：",listeners.length);
      listeners.forEach((cb) => cb(message)); 
    };
  
    socket.onerror = (error) => {
      console.error("WebSocketエラー:", error);
    };
  
    socket.onclose = () => {
      console.log("WebSocket接続が切断されました");
    };

  }
  return socket;
}
  

export function addMessageListener(callback: (msg: any) => void) {

  listeners.push(callback);

  console.log("addMessageListener:",listeners.length);
}

export function removeMessageListener(callback: (msg: any) => void){

  listeners = listeners.filter((cb) => cb !== callback)

  console.log("removeMessageListener:",listeners.length);
}

