import { useRouter } from "next/router";
import React from "react";
import { useState, useEffect, useRef } from "react";
import { createWebSocket } from "../utils/websocket";
import EmojiPicker from 'emoji-picker-react';
import Link from "next/link";

//import styles from "@/styles/Home.module.css";

interface Message {
  id: number;
  sender: number;
  sendername : string | null;
  type: "text" | "image" | "emoji";
  content: string;
  allread: boolean; // 既読状態を追跡するフラグ
  readcount: number;
}

const ChatRoom = () => {
  const router = useRouter();
  const [loggedInUser, setLoggedInUser] = useState<string | null>(null);
  const [loggedInUserid, setLoggedInUserid] = useState<number | null>(null);

  const { roomId } = router.query;

  const [messages, setMessages] = useState<Message[]>([]);
  const [message, setMessage] = useState("");
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null); // Refを使用
  const messagesEndRef = useRef<HTMLDivElement | null>(null);

  const [groupName, setGroupName] = useState<string | null>(null);
  const [socket, setSocket] = useState<WebSocket | null>(null);

  const [hoveredMessageId, setHoveredMessageId] = useState<number | null>(null);
  const [isOtherUserInRoom, setIsOtherUserInRoom] = useState(false);
  const isOtherUserInRoomRef = useRef(false);
  const hoverTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const [editingId, setEditingId] = useState<number | null>(null); // 編集中のメッセージID
  const [isEditing, setIsEditing] = useState(false);  
  const [editText, setEditText] = useState<string>(""); // 編集中の内容
  const [showEmojiPicker, setShowEmojiPicker] = useState(false); // 絵文字



  useEffect(() => {
    // 下までスクロール
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);


useEffect(() => {
  const setupChat = async () => {
    try {
      // --- ローカルストレージから取得 ---
      const token = localStorage.getItem("token");
      const username = localStorage.getItem("loggedInUser");
      const useridStr = localStorage.getItem("loggedInUserID");
      const roomName = localStorage.getItem("roomName");
      setGroupName(roomName);

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

      // Socket Open時のイベント
      ws.onopen = async () => {
        console.log("✅ WebSocket接続完了");

        // ✅ 自分の入室通知
        if (roomId) {
          const joinEvent = {
            type: "join",
            roomId: parseInt(roomId as string),
            userId: userid,
          };
          ws.send(JSON.stringify(joinEvent));
          console.log("🟢 入室通知送信:", joinEvent);
          setMessages((prev) => prev.map((msg) => ({ ...msg, isRead: true })));
        }
        

        // ✅ メッセージ履歴取得
        const res = await fetch(`http://localhost:8080/getRoomMessages?room_id=${roomId}`);
        // console.log("生データ：", res.json);
        const data = await res.json();
        console.log("😭",data.messages);
        //console.log("メッセージID：", data.messages[0]?.id);
        
        // console.log(JSON.stringify(data, null, 2));

        setMessages(data.messages); // BUG ← isRead が true になってる
        console.log("😭2",messages);

        // ✅ nullチェック追加！
        if (data && Array.isArray(data.messages)) {
          setMessages(data.messages);
        } else {
          setMessages([]); // nullや不正な値の場合は空配列
        }

        // とりあえずコメント
        // if (data && Array.isArray(data.messages)) {
        //   const formatted: Message[] = data.messages.map((msg: any) => ({
        //     type:"chat", // ✅ 自動判別でもOK
        //     id: msg.message_id,
        //     sender: msg.sender_id,
        //     sendername: msg.sender_name,
        //     content: msg.content || "(空メッセージ)",
        //     isRead: msg.is_read ?? false,
        //   }));
        //   console.log("🔍 formatted:", formatted);
        //   setMessages(formatted);
        // }


        // ✅ 一括既読更新（画面表示された履歴分）
        console.log("FFFFF");
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
          //setMessages((prev) => prev.map((msg) => ({ ...msg, allread: true })));  // この行で、allreadが全てtrueに上書きされる！
        }
      };

      // ✅ WebSocket受信処理
      ws.onmessage = async (event) => {
        try {
          const msg = JSON.parse(event.data);
          console.log("📩 WebSocket受信:", msg);

          //✅ user_joined メッセージは無視（または通知として別処理）
          if (msg.type === "user_joined") {
            console.log("👥 入室通知イベントを受信:", msg.userId);

            // ✅ 自分以外が入室してきたときに true にする
            if (Number(msg.userId) !== Number(userid)) {
              isOtherUserInRoomRef.current = true;
              setIsOtherUserInRoom(true);
              //console.log("✅ 他のユーザーが入室：isOtherUserInRoom = true");
              console.log("✅ isOtherUserInRoom = ",isOtherUserInRoom);
            }
            return;
          }

          // 新しいメッセージの既読情報の更新
          if (msg.type === "newreadmessage") {
            console.log("既読更新：", msg);
            console.log("ルームID既読msg.roomId：",msg.roomId);
            console.log("ルームID既読roomId：",roomId);
            if (String(msg.roomId) !== roomId) {
              console.log("ルームID既読：aaaaaaaaaaaa");
              return;
            }
            console.log("ルームID既読：bbbbbbbbbbbbb");

            interface SendMessages {
              room_id: number;
              message_id: number;
              readcount: number;
              allread: boolean;
            }

            // SendMessagesをMapに変換して高速アクセス
            const sendMap = new Map<number, SendMessages>();
            for (const sm of msg.newReadMessage) {
              sendMap.set(sm.message_id, sm);
            }
            console.log("sendMap：",sendMap);

            // messagesを上書きして新しい配列を返す
            setMessages((prevMessages) =>
              prevMessages.map(msglist => {
                //console.log("Messages.mapスタート");
                const readInfo = sendMap.get(msglist.id);
                if (readInfo) {
                  console.log("readInfo:", msglist.id, " > ", msglist.content, " > ", msglist.readcount);
                  return {
                    ...msglist,
                    allread: readInfo.allread,
                    readcount: readInfo.readcount
                  };
                }
                return msglist;
              })
            );

            return;
          }

          // ✅ 通常のチャットメッセージのみ以下を実行
          if (!msg.id || !msg.content || typeof msg.content !== "string") {
            console.warn("⚠️ 無効なチャットメッセージ:", msg);
            return;
          }        
          console.log("👤：",msg.sender, userid);

          if (Number(msg.sender) === Number(userid)) {
            console.log("☀️ 自分のメッセージなのでスキップ");
            return;
          }

          // ✅ 既読リクエスト（自分のメッセージは除外）
          if (Number(msg.sender) !== Number(userid)) {
            const res = await fetch(`http://localhost:8080/read`, {
              method: "POST",
              headers: {
                "Content-Type": "application/json",
                "Authorization": `Bearer ${token}`,
              },
              body: JSON.stringify({ login_id: userid, msg_id: msg.id }),
            });
            if (!res.ok) {
              throw new Error("未読一覧取得失敗");
            }

            const data = await res.json();
            console.log("PP：",data.data.MessageID);  // エラー、undefind
          } 


          // ✅ 表示に追加
          const newMessage: Message = {
            id: msg.id,
            sender: msg.sender,
            sendername: msg.sendername,
            type: msg.content.includes("/uploads/") ? "image" : "text", // ✅ 自動判別でもOK
            content: msg.content,
            allread: msg.read ?? false,
            readcount: msg.readcount,
          };
          setMessages((prev) => [...prev, newMessage]);
        } catch (err) {
          console.error("❌ WebSocket受信処理エラー:", err);
        };
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
}, [roomId]);

//console.log("😢：", messages[0]?.id); // undefined

  // onClickから呼ばれる
  // テキスト送信
  const handleSendMessage = async () => {
    console.log("xxxxxxxxxxxxxxxx:", messages);
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
      
      // 
      const isOnlyEmoji = /^[\p{Emoji}]{1}$/u.test(message.trim());

      const savedMessage: Message = {
        id: response.data.ID,
        sender: loggedInUserid ?? 0,
        sendername: loggedInUser,
        // type: selectedFile ? "image" : "text", // ✅ ファイルがある＝画像
        type: isOnlyEmoji ? "emoji" : (selectedFile ? "image" : "text"),  // ← 追加
        content: message.trim(),
        allread: false,
        readcount: 0,
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
      const res = await fetch("http://localhost:8080/sendFile", {
        method: "POST",
        body: formData,
        headers: {
          // Content-Typeを指定しない → formDataが勝手に解釈してくれる
        },
      });

      console.log("🔺レスポンス")
      if (!res.ok) {
        throw new Error("アップロード失敗");
      }


      const response = await res.json();
      console.log("アップロード成功: " + response);

      console.log("📨データ：", response);
      console.log("📨データ ID：", response.data.ID);
      const savedMessage: Message = {
        id: response.data.ID,
        sender: loggedInUserid ?? 0,
        sendername: loggedInUser,
        type: selectedFile ? "image" : "text", // ✅ ファイルがある＝画像
        content: response.image,
        allread: false,
        readcount: 0,
      };

      // WebSocket送信
      console.log("sockect：",savedMessage);
      if (socket) {
        socket.send(JSON.stringify(savedMessage));
      }

      setMessages((prev) => [...prev, savedMessage]);
      setMessage("");
      
      // ファイル選択をクリア
      setSelectedFile(null);
      if (fileInputRef.current) {
        fileInputRef.current.value = ""; // 実際に選択UIをクリア
      }

    } catch (error) {
      alert("アップロードエラー：" + error);
    }
    console.log("🔍 content:", messages); // タイプを変更 chat → image
  };

  //メッセージのリアクション、編集、削除
  type MessageAction = {
    id: string;
    text: string;
    isOwnMessage: boolean;
  };

  type ChatMessageProps = {
    messageaction: MessageAction;
    onUpdate: (id: string, newText: string) => void;
    onDelete: (id: string) => void;
  };

  // const ChatMessage: React.FC<{
  //   messageaction: MessageAction;
  //   onUpdate: (id: string, newTsxt: string) => void;
  //   onDelete: (id: string) => void;
  // }> = ({messageaction, onUpdate, onDelete}) => {
  //   // const [isEditing, setIsEditing] = useState(false);
  //   // const [editText, setEditText] = useState(messageaction.text);
  //   // const [hovered, setHovered] = useState(false);
  // }
  // console.log(ChatMessage);

  // const handleSave = () => {
  //   if(editText.trim()!==""){
  //     onUpdate(messageaction.id, editText);
  //     setIsEditing(false);
  //   }
  // };

  //const [hovered, setHovered] = useState(false);

  //リアクション
  const handleReact = (id: number) => {
    console.log("リアクション:", id);
  };
  
  // 編集    console.log("編集:", id);
  const handleEdit = async (id: number) => {
    const hoveredMessage = messages.find(msg => msg.id === hoveredMessageId);
    console.log("-----1：", hoveredMessage);

    if (editText.trim() === "") {
      setIsEditing(false);
      alert("メッセージを入力して下さい");
      console.log("-----2：", hoveredMessage);
      return;
    }
    try{
      const res = await fetch(`http://localhost:8080/editMessage?id=${id}`,{
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({content: editText}),
      });
      console.log("-----3：", hoveredMessage);

      if(!res.ok) throw new Error("編集失敗");

      setMessages((prev) =>
        prev.map((msg) => (msg.id === id ? { ...msg, content: editText } : msg))
      );
      setEditingId(null);
    } catch(error) {
      console.error("保存失敗", error);
      alert("メッセージの更新に失敗しました...")
    }
  }

  


  // 削除 と 取り消し //(msg)?
  const handleDelete = async (id: number) => {
    const hoveredMessage = messages.find(msg => msg.id === hoveredMessageId);
    console.log("-----：", hoveredMessageId);
    console.log("メッセージID📝：", hoveredMessage);
  
    console.log("削除：", id);
    const confirmed = window.confirm("このメッセージを削除しますか？");
    if (!confirmed) return;
    
    // 削除処理の実装へ
    try{
      const res = await fetch(`http://localhost:8080/deleteMessage?id=${id}`, { // id = message.id
        method: "DELETE",
      });
        if (!res.ok) {
          throw new Error("削除失敗");
        } else {
          alert("メッセージを削除しました");
        }
        // onDelete(id); // ローカル状態から削除
        // setMessages((prev) => prev.filter((msg) => msg.id !== id));

        // const deletedMessage = messages.find((msg) => msg.id === id);
        // // 1. 削除対象を取り除く
        // setMessages((prev) => prev.filter((msg) => msg.id !== id));
        // // 2. 削除ログメッセージを新しく追加
        // if (deletedMessage) {
        //   const logMsg = {
        //     id: Date.now(), // 仮のID
        //     sender: 0, // システムメッセージ的な扱い
        //     sendername: null,
        //     type: "text",
        //     content: `${deletedMessage.sendername ?? "ユーザー"}がメッセージを削除しました`,
        //     allread: true,
        //   };
          //setMessages((prev) => [...prev, logMsg]);

              // ✅ メッセージを「削除済み表示」に差し替える
          setMessages((prev) =>
            prev.map((msg) =>
              msg.id === id
                ? {
                    ...msg,
                    content: "（このメッセージは削除されました）",
                    type: "text", // 念のため
                  }
                : msg
            )
          );

          console.log(`🗑️ メッセージ${id}を削除しました`);
      
      } catch (err) {
        alert("削除できませんでした");
        console.error("削除エラー：", err);
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
          {messages.length >= 0 ? ( // messagesが空？
            messages.map((msg, index) => {
              const isMyMessage = String(msg.sender) === String(loggedInUserid);
              return (
                <div
                  key={`message-${index}-${msg.id}`}
                  style={{
                    position: "relative",
                    display: "flex",
                    flexDirection: "column",
                    alignItems: isMyMessage ? "flex-end" : "flex-start",
                    marginBottom: "20px",
                  }}
                  onMouseEnter={() => {
                    hoverTimeoutRef.current = setTimeout(() => {
                      setHoveredMessageId(msg.id);
                    }, 1000); // 1000ms待って表示
                  }}
                  onMouseLeave={() => {
                    if (hoverTimeoutRef.current) {
                      clearTimeout(hoverTimeoutRef.current);
                      hoverTimeoutRef.current = null;
                    }
                    setHoveredMessageId(null);
                  }}
                >
                  {/* ユーザー名（メッセージボックスの上） */}
                  <div
                    style={{
                      fontSize: "14px",
                      color: "#666",
                      fontWeight: 500,
                      marginBottom: "4px",
                      paddingLeft: isMyMessage ? undefined : "8px",
                      paddingRight: isMyMessage ? "8px" : undefined,
                      textAlign: isMyMessage ? "right" : "left",
                      width: "100%",
                    }}
                  >
                    {msg.sendername}
                  </div>
              
                  {/* 吹き出し（本文 or 画像） */}
                  <div
                    style={{
                      backgroundColor: isMyMessage ? "#dcf8c6" : "#ffffff",
                      color: "#333",
                      padding: "10px 14px 18px",
                      borderRadius: isMyMessage
                        ? "18px 18px 0 18px"
                        : "18px 18px 18px 0",
                      maxWidth: "60%",
                      boxShadow: "0 1px 4px rgba(0, 0, 0, 0.1)",
                      wordBreak: "break-word",
                      position: "relative",
                    }}
                  >
                    {editingId === msg.id ? (
                      <>
                        <input
                          value={editText}
                          onChange={(e) => setEditText(e.target.value)}
                          style={{
                            width: "100%",
                            padding: "8px",
                            fontSize: "16px",
                            border: "1px solid #ccc",
                            borderRadius: "8px",
                          }}
                          autoFocus
                        />
                          <div style={{ marginTop: "6px", display: "flex", gap: "10px" }}>
                            <button
                              onClick={() => handleEdit(msg.id)}
                              style={{ padding: "4px 10px", fontSize: "13px" }}
                            >
                              保存
                            </button>
                            <button
                              onClick={() => setEditingId(null)}
                              style={{ padding: "4px 10px", fontSize: "13px", color: "#777" }}
                            >
                              キャンセル
                            </button>
                          </div>
                        </>
                      ) : (
                        <>
                      
                    {/* 本文 or 画像 or 絵文字 */}
                    {msg.type === "emoji" ? (
                      <div
                        style={{
                          fontSize: "120px",          // 大きく
                          textAlign: "center",
                          padding: "10px 0",
                          lineHeight: "1",
                        }}
                      >
                        {msg.content}
                      </div>
                    ) : (
                    msg.content.startsWith("http") &&
                      msg.content.match(/\.(jpg|jpeg|png|gif|webp)(\?.*)?$/i) ? (
                      <img
                        src={msg.content}
                        alt="画像"
                        style={{
                          maxWidth: "70%",
                          borderRadius: "10px",
                          border: "1px solid #ccc",
                          marginTop: "4px",
                        }}
                      />
                    ) : (
                      <div
                        style={{
                          fontSize: "17px",
                          lineHeight: "1.6",
                          whiteSpace: "pre-wrap",
                          color: msg.content === "（このメッセージは削除されました）" ? "#888" : "#222",
                          fontStyle: msg.content === "（このメッセージは削除されました）" ? "italic" : "normal",
                        }}
                      >
                        {msg.id} : {msg.content}
                      </div>
                    ))}
                  </>
                )}
                    {/* 既読 */}
                    {/* {msg.allread && isMyMessage && isOtherUserInRoomRef.current && ( */}
                    {/* {isMyMessage && ( */}
                      <div
                        style={{
                          fontSize: "11px",
                          color: "green",
                          position: "absolute",
                          bottom: "1px",
                          right: "10px",
                        }}
                      >
                        既読( {msg.readcount} )
                      </div>
                    {/* )} */}
                    </div>
              
                  {/* ホバーメニュー */}
                  {hoveredMessageId === msg.id && (
                    <div
                      style={{
                        position: "absolute",
                        bottom: "-30px",
                        right: isMyMessage ? "0" : "auto",
                        left: isMyMessage ? "auto" : "0",
                        backgroundColor: "#fff",
                        border: "1px solid #ccc",
                        borderRadius: "8px",
                        padding: "6px 10px",
                        boxShadow: "0 2px 4px rgba(0, 0, 0, 0.1)",
                        display: "flex",
                        gap: "8px",
                        zIndex: 10,
                      }}
                    >
                      {isMyMessage ? (
                        <>
                          <span 
                          style={{
                            fontSize: "13px",
                            cursor: "pointer",
                          }}
                          //  onClick={() => handleEdit(msg.id)}>編集</span>
                          onClick={() => {
                            setEditingId(msg.id);
                            setEditText(msg.content);
                          }}
                          >編集</span>

                          <span
                          style={{
                            fontSize: "13px",
                          }}
                          onClick={() => handleDelete(msg.id)}>削除</span>
                        </>
                      ) : (
                        <span 
                        style={{
                          fontSize: "13px",
                        }}
                        onClick={() => handleReact(msg.id)}>👍</span>
                      )}
                    </div>
                  )}
              
                  <div ref={messagesEndRef} />
                </div>
              );
              }
            )
          ) : (
            <p>メッセージがありません</p>
          )}
        </div>
                <div style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          marginTop: "15px",
          position: "relative",
          width: "100%"
        }}>
          {/* 左下：絵文字とファイル */}
          <div style={{ display: "flex", gap: "10px" }}>
            <button onClick={() => setShowEmojiPicker(prev => !prev)}>😊</button>
            <input
              type="file"
              onChange={handleFileChange}
              ref={fileInputRef}
              style={{ fontSize: "13px" }}
            />
          </div>

          {/* 中央：入力欄 */}
          <input
            type="text"
            placeholder="メッセージを入力"
            value={selectedFile ? selectedFile.name : message}
            onChange={(e) => {
              setMessage(e.target.value);
              setSelectedFile(null); // テキスト入力されたらファイルリセット
            }}
            style={{
              flex: 1,
              margin: "0 10px",
              padding: "16px",
              borderRadius: "30px",
              border: "2px solid #ccc"
            }}
          />

          {/* 右：送信 */}
          <button onClick={() => {
            if (selectedFile) {
              handleSubmit(); // ファイル送信
            } else {
              handleSendMessage(); // テキスト送信
            }
          }} style={{
            fontSize: "15px",
            backgroundColor: "#388e3c",
            color: "#fff",
            padding: "10px 25px",
            borderRadius: "30px",
            border: "none",
            cursor: "pointer"
          }}>
            送信
          </button>

          {/* Emoji Picker ポップアップ */}
          {showEmojiPicker && (
            <div style={{
              position: "absolute",
              bottom: "60px",
              left: "0px",
              zIndex: 100
            }}>
              <EmojiPicker
                onEmojiClick={(emojiData) => {
                  setMessage(prev => prev + emojiData.emoji);
                  setShowEmojiPicker(false);
                }}
              />
            </div>
          )}
        </div>
          <footer style={{ marginTop: "20px", textAlign: "center" }}>
            <Link href="/roomSelect" style={{ color: "#388e3c", marginRight: "10px" }}>← 戻る</Link>
          </footer>
      </div>
      </div>

  );
};

export default ChatRoom;
