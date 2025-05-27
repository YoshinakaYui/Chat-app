// // WebSocketContext.tsx
// import { createContext, useContext, useEffect, useRef } from "react";

// const WebSocketContext = createContext<WebSocket | null>(null);

// export const WebSocketProvider = ({ children }: { children: React.ReactNode }) => {
//   const socketRef = useRef<WebSocket | null>(null);

//   useEffect(() => {
//     socketRef.current = new WebSocket("ws://localhost:8080/ws");
//     return () => {
//       socketRef.current?.close();
//     };
//   }, []);

//   return (
//     <WebSocketContext.Provider value={socketRef.current}>
//       {children}
//     </WebSocketContext.Provider>
//   );
// };

// export const useWebSocket = () => useContext(WebSocketContext);
