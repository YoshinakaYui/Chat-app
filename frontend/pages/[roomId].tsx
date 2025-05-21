import { useRouter } from "next/router";
import { useState, useEffect, useRef } from "react";
import { createWebSocket } from "../utils/websocket";
import Link from "next/link";

//import styles from "@/styles/Home.module.css";

interface Message {
  id: number;
  sender: number;
  sendername : string | null;
  content: string;
  isRead: boolean; // 既読状態を追跡するフラグ
}

const ChatRoom = () => {
  const router = useRouter();
  const { roomId } = router.query;
  const [messages, setMessages] = useState<Message[]>([]);
  const [message, setMessage] = useState("");
  const [loggedInUser, setLoggedInUser] = useState<string | null>(null);
  const [loggedInUserid, setLoggedInUserid] = useState<number | null>(null);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null); // Refを使用
  const [groupName, setGroupName] = useState<string | null>(null);
  const [socket, setSocket] = useState<WebSocket | null>(null);
  const messagesEndRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    // 下までスクロール
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  // useEffect(() => {
  //   const fetchMessages = async () => {
  //     try {
  //       const token = localStorage.getItem("token");
  //       if (!token) {
  //         alert("ログインされていません");
  //         router.push("/top");
  //         return;
  //       }
  //       console.log("🍏：",roomId, loggedInUserid,socket);

  //       // ルームの入室を通知
  //       if (roomId && loggedInUserid && socket) {
  //         const joinEvent = {
  //           type: "join",
  //           roomId: parseInt(roomId as string),
  //           userId: loggedInUserid,
  //         };
  //         console.log("🟢：", joinEvent);
  //         socket.send(JSON.stringify(joinEvent));
  //         console.log("🟢 入室通知を送信:", joinEvent);
  //       }

  //         // クライアントサイドでのみ実行するためのチェック
  //         // if (typeof window !== "undefined") {
  //         //   const storedRoomName = localStorage.getItem("roomName");
  //         //   if (storedRoomName) {
  //         //     setGroupName(storedRoomName);
  //         //   } else {
  //         //     console.warn("ルーム名が見つかりません");
  //         //   }
  //         // }
  //       const res = await fetch(`http://localhost:8080/getRoomMessages?room_id=${roomId}`);
  //       if (!res.ok) {
  //         throw new Error(`HTTPエラー: ${res.status}`);
  //       }

  //       console.log("ルームID：",roomId)
  //       const data = await res.json();
  //       if (data && Array.isArray(data.messages)) {

  //         console.log("✉️データ：",data)

  //         const formattedMessages: Message[] = data.messages.map((msg: any) => ({
  //           id: msg.message_id,
  //           sender: msg.sender_id.toString(),
  //           sendername : msg.sender_name,
  //           content: msg.content,
  //         }));
  //         setMessages(formattedMessages);
  //       }

  //       // 未読の更新
  //       try{
  //         const res = await fetch(`http://localhost:8080/updataUnReadMessage`, {
  //           method: "POST",
  //           headers: {
  //             "Content-Type": "application/json",
  //             "Authorization": `Bearer ${token}`,
  //           },
  //           body: JSON.stringify({login_id: parseInt(loggedInUserid as string), room_id:parseInt(roomId as string)})
  //         });
  //         const result = await res.json();
  //         if (res.ok) {
  //           console.log("📥 read レスポンス:", result);
  //           setMessages((prev) =>
  //             prev.map((msg) => ({ ...msg, isRead: true })) // ✅ すべて既読に
  //           );
  //         }

  //       } catch {
  //         console.log("失敗");
  //       }

  //     } catch (err) {
  //       console.error("メッセージ取得エラー:", err);
  //       setMessages([]);
  //     }
  //   };
  //   console.log("🟩",setMessages);

  //   const loggedInUsername = localStorage.getItem("loggedInUser");
  //   const loggedInUserid = localStorage.getItem("loggedInUserID");
  //   if (loggedInUsername) setLoggedInUser(loggedInUsername);
  //   if (loggedInUserid) setLoggedInUserid(parseInt(loggedInUserid ?? "0",10));

  //   if (roomId) fetchMessages();
  // }, [roomId]);

  // // WebSocket利用準備 & socketからの受信
  // useEffect(() => {
  //   try{
  //     const token = localStorage.getItem("token");
  //     const ws = createWebSocket(async (msg) => {
  //       console.log("📩 WebSocketで受信:", msg);
  //       console.log("🧪 msg.senderid:", msg.senderid, "typeof:", typeof msg.senderid);
  //       console.log("🧪 loggedInUserid:", loggedInUserid, "typeof:", typeof loggedInUserid);
  //       console.log("🧪 parsed:", loggedInUserid);
        
  //       console.log("☀️：", msg.sendername);

  //     if (!msg.id) {
  //       console.warn("⚠️", msg.id, "undefined");
  //       return;
  //     }
  //     console.log("😺",msg);

  //       try{
  //         const res = await fetch(`http://localhost:8080/read`, {
  //           method: "POST",
  //           headers: {
  //             "Content-Type": "application/json",
  //             "Authorization": `Bearer ${token}`,
  //           },
  //           body: JSON.stringify({login_id: loggedInUserid, msg_id: msg.id})
  //         });
  //         //const response = await res.json();
  //         const result = await res.json();
  //         console.log("📥 read レスポンス:", result);

  //       } catch {
  //         console.log("失敗");
  //       }


  //     // 自分自身が送信したメッセージなら、WebSocketからの受信はスキップ
  //     if (String(msg.sender) === String(loggedInUserid)) {
  //       console.log("☀️ スキップ：自分が送ったメッセージ");
  //       return; // 表示しない
  //     }

  //     const newMessage: Message = {
  //       id: msg.id,
  //       sender: msg.sender,
  //       sendername: msg.sendername,
  //       content: msg.content,
  //       isRead: msg.read,
  //     };

  //     setMessages((prev) => [...prev, newMessage]);
  //   });
  //   setSocket(ws);

  //   return () => ws.close();
  // }catch (err){
  //   console.error("❌ useEffect 全体エラー:", err);
  // }
  // }, [loggedInUserid]);

  //メッセージ取得：入室時


// 合体版↓
//   useEffect(() => {
//   const fetchAndSetup = async () => {
//     try {
//       const token = localStorage.getItem("token");
//       const loggedInUsername = localStorage.getItem("loggedInUser");
//       const loggedInUseridStr = localStorage.getItem("loggedInUserID");

//       if (!token || !loggedInUseridStr) {
//         alert("ログインされていません");
//         router.push("/top");
//         return;
//       }

//       setLoggedInUser(loggedInUsername ?? "");
//       const loggedInUseridNum = parseInt(loggedInUseridStr, 10);
//       setLoggedInUserid(loggedInUseridNum);

//       // WebSocket 初期化
//       const ws = createWebSocket(async (msg) => {
//         console.log("📩 WebSocket受信:", msg);

//         // 既読リクエスト
//         try {
//           const res = await fetch(`http://localhost:8080/read`, {
//             method: "POST",
//             headers: {
//               "Content-Type": "application/json",
//               "Authorization": `Bearer ${token}`,
//             },
//             body: JSON.stringify({ login_id: loggedInUseridNum, msg_id: msg.id }),
//           });
//           const result = await res.json();
//           console.log("📥 read レスポンス:", result);
//         } catch {
//           console.log("❌ 既読登録失敗");
//         }

//         if (String(msg.sender) === String(loggedInUseridNum)) {
//           console.log("☀️ スキップ：自分のメッセージ");
//           return;
//         }

//         const newMessage: Message = {
//           id: msg.id,
//           sender: msg.sender,
//           sendername: msg.sendername,
//           content: msg.content,
//           isRead: msg.read,
//         };

//         setMessages((prev) => [...prev, newMessage]);
//       });

//       setSocket(ws);

//       // 🎯 joinイベント送信（socket生成後）
//       if (roomId) {
//         const joinEvent = {
//           type: "join",
//           roomId: parseInt(roomId as string),
//           userId: loggedInUseridNum,
//         };
//         console.log("🟢 入室通知:", joinEvent);
//         ws.onopen = () => {
//           ws.send(JSON.stringify(joinEvent));
//           console.log("🟢 join送信完了");
//         };
//       }

//       // 📥 メッセージ取得
//       const res = await fetch(`http://localhost:8080/getRoomMessages?room_id=${roomId}`);
//       const data = await res.json();

//       if (data && Array.isArray(data.messages)) {
//         const formattedMessages: Message[] = data.messages.map((msg: any) => ({
//           id: msg.message_id,
//           sender: msg.sender_id.toString(),
//           sendername: msg.sender_name,
//           content: msg.content,
//           isRead: msg.is_read ?? false,
//         }));
//         setMessages(formattedMessages);
//       }

//       // 📘 未読→既読処理
//       const markRes = await fetch(`http://localhost:8080/updataUnReadMessage`, {
//         method: "POST",
//         headers: {
//           "Content-Type": "application/json",
//           "Authorization": `Bearer ${token}`,
//         },
//         body: JSON.stringify({
//           login_id: loggedInUseridNum,
//           room_id: parseInt(roomId as string),
//         }),
//       });
//       const markResult = await markRes.json();
//       if (markRes.ok) {
//         console.log("✅ 既読更新:", markResult);
//         setMessages((prev) => prev.map((msg) => ({ ...msg, isRead: true })));
//       }

//     } catch (err) {
//       console.error("❌ 全体エラー:", err);
//       setMessages([]);
//     }
//   };

//   if (roomId) {
//     fetchAndSetup();
//   }

//   return () => {
//     if (socket) {
//       socket.close();
//       console.log("👋 WebSocket切断");
//     }
//   };
// }, [roomId]);

useEffect(() => {
  const setupChat = async () => {
    try {
      // --- ローカルストレージから取得 ---
      const token = localStorage.getItem("token");
      const username = localStorage.getItem("loggedInUser");
      const useridStr = localStorage.getItem("loggedInUserID");

      if (!token || !useridStr) {
        alert("ログインされていません");
        router.push("/top");
        return;
      }

      const userid = parseInt(useridStr, 10);
      setLoggedInUser(username ?? "");
      setLoggedInUserid(userid);

      // --- WebSocket初期化 ---
      const ws = new WebSocket("ws://localhost:8080/ws");

      ws.onopen = async () => {
        console.log("✅ WebSocket接続完了");

        // ✅ 入室通知
        if (roomId) {
          const joinEvent = {
            type: "join",
            roomId: parseInt(roomId as string),
            userId: userid,
          };
          ws.send(JSON.stringify(joinEvent));
          console.log("🟢 入室通知送信:", joinEvent);
        }

        // ✅ メッセージ履歴取得
        const res = await fetch(`http://localhost:8080/getRoomMessages?room_id=${roomId}`);
        const data = await res.json();

        if (data && Array.isArray(data.messages)) {
          const formatted: Message[] = data.messages.map((msg: any) => ({
            type: "chat",
            id: msg.message_id,
            sender: msg.sender_id.toString(),
            sendername: msg.sender_name,
            content: msg.content || "(空メッセージ)",
            isRead: msg.is_read ?? false,
          }));
          setMessages(formatted);
        }

        // ✅ 一括既読更新（画面表示された履歴分）
        const markRes = await fetch(`http://localhost:8080/updataUnReadMessage`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "Authorization": `Bearer ${token}`,
          },
          body: JSON.stringify({
            login_id: userid,
            room_id: parseInt(roomId as string),
          }),
        });
        const markResult = await markRes.json();
        if (markRes.ok) {
          console.log("✅ 履歴既読化成功:", markResult);
          setMessages((prev) => prev.map((msg) => ({ ...msg, isRead: true })));
        }
      };

      // ✅ WebSocket受信処理
      ws.onmessage = async (event) => {
        try {
          const msg = JSON.parse(event.data);
          console.log("📩 WebSocket受信:", msg);

        // ✅ user_joined メッセージは無視（または通知として別処理）
        if (msg.type === "user_joined") {
          console.log("👥 入室通知イベントを受信:", msg.userId);
          return;
        }

        // ✅ 通常のチャットメッセージのみ以下を実行
        if (!msg.id || !msg.content || typeof msg.content !== "string") {
          console.warn("⚠️ 無効なチャットメッセージ:", msg);
          return;
        }

        

        // if (Number(msg.sender) === Number(userid)) {
        //   console.log("☀️ 自分のメッセージなのでスキップ");
        //   return;
        // }

          // ✅ 既読リクエスト（自分のメッセージは除外）
          if (String(msg.sender) !== String(userid)) {
            await fetch(`http://localhost:8080/read`, {
              method: "POST",
              headers: {
                "Content-Type": "application/json",
                "Authorization": `Bearer ${token}`,
              },
              body: JSON.stringify({ login_id: userid, msg_id: msg.id }),
            });
          } 
          // else {
          //   console.log("☀️ 自分のメッセージは既読処理スキップ");
          // }

          // ✅ 表示追加
    // ✅ 表示に追加
    const newMessage: Message = {
      id: msg.id,
      sender: msg.sender,
      sendername: msg.sendername,
      content: msg.content,
      isRead: msg.read ?? false,
    };
          setMessages((prev) => [...prev, newMessage]);
        } catch (err) {
          console.error("❌ WebSocket受信処理エラー:", err);
        }
      };

      setSocket(ws); // socketステートにセット

    } catch (err) {
      console.error("❌ チャット初期化エラー:", err);
      setMessages([]);
    }
  };

  if (roomId) {
    setupChat();
  }

  // ❌ WebSocket切断はログアウト時のみ → ここでは close() しない
}, [roomId]);


  const handleSendMessage = async () => {
    if (!message.trim()) {
      alert("メッセージを入力してください");
      return;
    }

    try {
      const newMessage = {
        roomid: parseInt(roomId as string, 10),
        senderid: loggedInUserid,
        content: message.trim(),
      };

      const res = await fetch("http://localhost:8080/message", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(newMessage),
      });

      if (!res.ok) {
        throw new Error("メッセージ送信失敗");
      }

      const response = await res.json();
      console.log("📨データ：", response);
      console.log("📨データ ID：", response.data.ID);
      const savedMessage: Message = {
        id: response.data.ID,
        sender: loggedInUserid ?? 0,
        sendername: loggedInUser,
        content: message.trim(),
        isRead: true  // ✅ 自分が送ったメッセージなので既読扱い
      };

      // WebSocket送信
      console.log("sockect：",savedMessage);
      if (socket) {
        socket.send(JSON.stringify(savedMessage));
      }

      setMessages((prev) => [...prev, savedMessage]);
      setMessage("");
    } catch (err) {
      alert("メッセージ送信エラー");
      console.error("送信エラー:", err);
    }
  };

  // 新しい人が入室したかどうか
  // senderが自分じゃない場合は、既読カウントしない

  // ファイル選択
  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.files && event.target.files.length > 0) {
      setSelectedFile(event.target.files[0]);
    }
  };

  // ファイル送信
  const handleSubmit = async () => {
    if (!selectedFile) {
      alert("ファイルを選択してください");
      return;
    }

    console.log(loggedInUserid);

    const formData = new FormData();
    formData.append("file",selectedFile);
    formData.append("senderID",String(loggedInUserid));
    formData.append("roomID",String(roomId));

    try {
      const response = await fetch("http://localhost:8080/sendFile", {
        method: "POST",
        body: formData,
        headers: {
          // Content-Typeを指定しない → formDataが勝手に解釈してくれる
        },
      });

      console.log("🔺レスポンス")
      if (!response.ok) {
        throw new Error("アップロード失敗");
      }

      const data = await response.text();
      //const fileURL = data.image;
      console.log("ファイルレスポンス：");
      alert("アップロード成功: " + data);

      // ファイル選択をクリア
      setSelectedFile(null);
      if (fileInputRef.current) {
        fileInputRef.current.value = ""; // 実際に選択UIをクリア
      }
    } catch (error) {
      alert("アップロードエラー：" + error);
    }
  };


  return (
    <div style={{
      background: "linear-gradient(180deg, #e8f5e9, #fffde7)",
      minHeight: "100vh",
      height: "100vh",
      overflow: "hidden",
      display: "flex",
      flexDirection: "column",
      justifyContent: "center",
      alignItems: "center"
    }}>
      <div style={{
        backgroundColor: "#ffffff",
        padding: "40px",
        borderRadius: "30px",
        boxShadow: "0px 4px 8px rgba(0,0,0,0.1)",
        width: "90%",
        maxWidth: "1000px",
        textAlign: "center"
      }}>
        <h2 style={{ color: "#388e3c", marginBottom: "15px" }}>ルーム：{groupName ? groupName : "ルーム名がありません"}</h2>
        <div style={{ maxHeight: "500px", overflowY: "scroll", marginBottom: "15px" }}>
          {messages.length > 0 ? (
            messages.map((msg, index) => (
              <p key={`message-${index}-${msg.id}`} style={{
                padding: "10px",
                borderRadius: "10px",
                margin: "10px 0",
                textAlign: String(msg.sender) === String(loggedInUserid) ? "right" : "left",
                alignSelf: msg.sender === loggedInUserid ? "flex-end" : "flex-start",
                maxWidth: "90%"
              }}>
                  <strong>{msg.sendername}:</strong> {msg.content}
                  {/* {msg.sender === loggedInUserid && msg.isRead && ( */}
                  {msg.isRead && (
                  <span style={{ fontSize: "12px", color: "green", marginLeft: "10px" }}>
                    （既読）
                  </span>
                )}
                <div ref={messagesEndRef} />
              </p>
            ))
          ) : (
            <p>メッセージがありません</p>
          )}
        </div>
        <div style={{ display: "flex", gap: "10px" }}>
          <div>
            <input 
              type="file"
              onChange={handleFileChange}
              ref={fileInputRef} // Refを設定  
            />
            <button onClick={handleSubmit}>アップロード</button>
          </div>
          <input
            type="text"
            placeholder="メッセージを入力"
            value={message}
            onChange={(e) => setMessage(e.target.value)}
            style={{ flex: 1, padding: "20px", borderRadius: "30px", border: "2px solid #ccc" }}
          />
          <button onClick={handleSendMessage} style={{
            backgroundColor: "#388e3c",
            color: "#fff",
            padding: "10px 30px",
            borderRadius: "30px",
            border: "none",
            cursor: "pointer",
            transition: "all 0.3s"
          }}>送信</button>
        </div>
          <footer style={{ marginTop: "20px", textAlign: "center" }}>
            <Link href="/roomSelect" style={{ color: "#388e3c", marginRight: "10px" }}>戻る →</Link>
          </footer>
      </div>
    </div>
  );
};

export default ChatRoom;
