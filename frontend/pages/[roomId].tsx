import { useRouter } from "next/router";
import { useState, useEffect, useRef } from "react";
import Link from "next/link";

//import styles from "@/styles/Home.module.css";

interface Message {
  id: number;
  sender: string;
  sendername : string | null;
  content: string;
}

const ChatRoom = () => {
  const router = useRouter();
  const { roomId } = router.query;
  const [messages, setMessages] = useState<Message[]>([]);
  const [message, setMessage] = useState("");
  const [loggedInUser, setLoggedInUser] = useState<string | null>(null);
  const [loggedInUserid, setLoggedInUserid] = useState<string | null>(null);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null); // Refを使用
  const [groupName, setGroupName] = useState<string | null>(null);

  //console.log(localStorage)

  // メッセージ送信
  useEffect(() => {
    const fetchMessages = async () => {
      try {
            // クライアントサイドでのみ実行するためのチェック
          console.log("🟢",localStorage)
          if (typeof window !== "undefined") {
            const storedRoomName = localStorage.getItem("roomName");
            if (storedRoomName) {
              setGroupName(storedRoomName);
            } else {
              console.warn("ルーム名が見つかりません");
            }
          }
        const res = await fetch(`http://localhost:8080/getRoomMessages?room_id=${roomId}`);
        if (!res.ok) {
          throw new Error(`HTTPエラー: ${res.status}`);
        }

        const data = await res.json();
        if (data && Array.isArray(data.messages)) {

          console.log("データ：",data)

          const formattedMessages: Message[] = data.messages.map((msg: any) => ({
            id: msg.message_id,
            sender: msg.sender_id.toString(),
            sendername : msg.sender_name,
            content: msg.content,
          }));
          setMessages(formattedMessages);
        }
      } catch (err) {
        console.error("メッセージ取得エラー:", err);
        setMessages([]);
      }
    };

    const loggedInUsername = localStorage.getItem("loggedInUser");
    const loggedInUserid = localStorage.getItem("loggedInUserID");
    if (loggedInUsername) setLoggedInUser(loggedInUsername);
    if (loggedInUserid) setLoggedInUserid(loggedInUserid);

    if (roomId) fetchMessages();
  }, [roomId]);

  // メッセージ処理
  const handleSendMessage = async () => {
    if (!message.trim()) {
      alert("メッセージを入力してください");
      return;
    }

    try {
      const newMessage = {
        roomid: parseInt(roomId as string, 10),
        senderid: parseInt(loggedInUserid || "0", 10),
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

      const data = await res.json();
      const savedMessage: Message = {
        id: data.id,
        sender: loggedInUserid || "未ログイン",
        sendername: loggedInUser,
        content: message.trim(),
      };

      setMessages((prev) => [...prev, savedMessage]);
      setMessage("");
    } catch (err) {
      alert("メッセージ送信エラー");
      console.error("送信エラー:", err);
    }
  };

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

    /*
    roomid: parseInt(roomId as string, 10),
    senderid: parseInt(loggedInUserid || "0", 10),
    */

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

      const result = await response.text();
      alert("アップロード成功: " + result);

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
      display: "flex",
      flexDirection: "column",
      justifyContent: "center",
      alignItems: "center"
    }}>
      <div style={{
        backgroundColor: "#ffffff",
        padding: "60px",
        borderRadius: "20px",
        boxShadow: "0px 4px 8px rgba(0,0,0,0.1)",
        width: "90%",
        maxWidth: "1000px",
        textAlign: "center"
      }}>
        <h2 style={{ color: "#388e3c", marginBottom: "15px" }}>ルーム：{groupName ? groupName : "ルーム名がありません"}</h2>
        <div style={{ maxHeight: "500px", overflowY: "auto", marginBottom: "15px" }}>
          {messages.length > 0 ? (
            messages.map((msg, index) => (
              <p key={`message-${index}-${msg.id}`} style={{
                padding: "10px",
                borderRadius: "10px",
                margin: "10px 0",
                textAlign: msg.sender === loggedInUserid ? "right" : "left",
                alignSelf: msg.sender === loggedInUserid ? "flex-end" : "flex-start",
                maxWidth: "90%"
              }}>
                <strong>{msg.sendername}:</strong> {msg.content}
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
