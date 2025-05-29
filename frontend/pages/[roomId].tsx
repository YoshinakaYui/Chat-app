import { useRouter } from "next/router";
import React from "react";
import { useState, useEffect, useRef } from "react";
import EmojiPicker from 'emoji-picker-react';
import Link from "next/link";
import { connectWebSocket, addMessageListener, removeMessageListener } from "../utils/websocket";

interface User {
  id: number;
  username: string;
}
interface Message {
  id: number;
  sender: number;
  sendername : string | null;
  type: "text" | "image";
  content: string;
  allread: boolean; // 既読状態を追跡するフラグ
  readcount: number;
  reaction?: string | null;
}

const ChatRoom = () => {
  const router = useRouter();
  const [loggedInUser, setLoggedInUser] = useState<string | null>(null);
  const [loggedInUserid, setLoggedInUserid] = useState<number | null>(null);
  const [isGroup, setIsGroup] = useState<number | null>(null);

  const { roomId } = router.query;

  const [messages, setMessages] = useState<Message[]>([]);
  const [message, setMessage] = useState("");
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null); // Refを使用
  const messagesEndRef = useRef<HTMLDivElement | null>(null);

  const [groupName, setGroupName] = useState<string | null>(null);
  const [currentRoomId, setCurrentRoomId] = useState<number | null>(null);
  const currentRoomIdRef = useRef<number | null>(null);

  const [hoveredMessageId, setHoveredMessageId] = useState<number | null>(null);
  const [isOtherUserInRoom, setIsOtherUserInRoom] = useState(false);
  const isOtherUserInRoomRef = useRef(false);
  const hoverTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const [editingId, setEditingId] = useState<number | null>(null); // 編集中のメッセージID
  const [isEditing, setIsEditing] = useState(false);  
  const [editText, setEditText] = useState<string>(""); // 編集中の内容
  const [showEmojiPicker, setShowEmojiPicker] = useState(false); // 絵文字
  const [showMentionList, setShowMentionList] = useState(false);

  const [selectedUsers, setSelectedUsers] = useState<number[]>([]);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isPersonalModalOpen, setIsPersonalModalOpen] = useState(false);

  const [members, setMembers] = useState<User[]>([]);
  const [notMembers, setNotMembers] = useState<User[]>([]);



    // メッセージを下までスクロール
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  // currentRoomIdを更新
  useEffect(() => {
    console.log("currentRoomId が変化した：", currentRoomId);
    currentRoomIdRef.current = currentRoomId;
  }, [currentRoomId]);

  // メッセージ履歴取得
  useEffect(() => {
    const setupChat = async () => {
      console.log("setupChat開始")
      try {
        // --- ローカルストレージから取得 ---
        const token = localStorage.getItem("token");
        const username = localStorage.getItem("loggedInUser");
        const useridStr = localStorage.getItem("loggedInUserID");
        const is_group = localStorage.getItem("is_group");
        const roomName = localStorage.getItem("roomName");
        const i_roomId = parseInt(roomId as string);
        console.log("i_roomId：",i_roomId);

        setIsGroup(parseInt(is_group?? "",10));
        setCurrentRoomId(i_roomId);
        console.log("currentRoomId：", currentRoomId);

        setGroupName(roomName);

        if (!token || !useridStr) {
          alert("ログインされていません");
          router.push("/top");
          return;
        }

        const userid = parseInt(useridStr, 10);
        setLoggedInUser(username ?? "");
        setLoggedInUserid(userid);

        console.log("✅ WebSocket接続完了");

        // ✅ 自分の入室通知
        if (roomId) {
          const joinEvent = {
            type: "join",
            roomId: parseInt(roomId as string),
            userId: userid,
          };
          console.log("🟢 入室通知送信:", joinEvent);
          setMessages((prev) => prev.map((msg) => ({ ...msg, isRead: true })));
        }
        

        // ✅ メッセージ履歴取得
        console.log("userid：", userid);
        const res = await fetch(`http://localhost:8080/getRoomMessages?room_id=${roomId}`,{
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
            "Authorization": `Bearer ${token}`,
          },
          body: JSON.stringify({ login_id: userid}),
        });

        const data = await res.json();
        setMessages(data.messages);

        // ✅ nullチェック追加！
        if (data && Array.isArray(data.messages)) {
          setMessages(data.messages);
        } else {
          setMessages([]); // nullや不正な値の場合は空配列
        }

        // ✅ 一括既読更新（画面表示された履歴分）
        console.log("FFFFF");
        const markRes = await fetch(`http://localhost:8080/updateUnReadMessage`, {
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
        }

      } catch (err) {
        console.error("❌ チャット初期化エラー:", err);
        setMessages([]);
      }
    };
    if (roomId) {
      setupChat();
    }

    // ✅ クリーンアップ処理で WebSocket を確実に閉じる
    return () => {
      // 離脱時はnullにする
      console.log("roomid clear.")
      setCurrentRoomId(null);
      currentRoomIdRef.current = null;
    };
  }, [roomId]);

  // メッセージ受け取り
  useEffect(() => {
    connectWebSocket();
    const token = localStorage.getItem("token");
    const useridStr = localStorage.getItem("loggedInUserID");
    const i_roomId = parseInt(roomId as string);
    console.log("i_roomId：",i_roomId);
    const userid = parseInt(useridStr ?? "",10);

    const handleMessage = async (msg: any) => {
      try {
        const roomId = Number(msg.room_id);
        const currentRoomId = Number(currentRoomIdRef.current);
        console.log("room_id value:", msg.room_id, "type:", typeof msg.room_id);
        console.log("currentRoomIdRef.current value:", currentRoomIdRef.current, "type:", typeof currentRoomIdRef.current);
        
        console.log("msg.room_id:", roomId, "currentRoomIdRef.current:",currentRoomId )


        if (parseInt(msg.room_id) !== currentRoomIdRef.current){
          console.log("msg.room_id：", roomId);
          console.log("currentRoomId：", currentRoomId);
          console.log("ルームIDが違います");
          return;
        }

        //✅ user_joined メッセージは無視（または通知として別処理）
        if (msg.type === "user_joined") {
          console.log("👥 入室通知イベントを受信:", msg.userId);

          // ✅ 自分以外が入室してきたときに true にする
          if (Number(msg.userId) !== Number(userid)) {
            isOtherUserInRoomRef.current = true;
            setIsOtherUserInRoom(true);
            console.log("✅ isOtherUserInRoom = ",isOtherUserInRoom);
          }
          return;
        }

        // 新しいメッセージの既読情報の更新
        if (msg.type === "newreadmessage") {
          console.log("既読更新：", msg);

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

          // messagesを上書きして新しい配列を返す
          setMessages((prevMessages) =>
            prevMessages.map(msglist => {
              const readInfo = sendMap.get(msglist.id);
              if (readInfo) {
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

        if(msg.type === "updateMessage"){
          if (String(loggedInUserid) === String(msg.messageid)){
          }
          console.log("受信したmsg:", msg);

          console.log("編集、削除を共有")
          setMessages((prevMessages) =>
            prevMessages.map(msglist => {
              console.log("Messages.mapスタート");
              if(msglist.id === msg.messageid){
                return{
                  ...msglist,
                  content: msg.content
                }
              }
              return msglist;
            })
          );
          return
        }

        // リアクション

        if (msg.type === "reaction") {
          console.log("リアクション受信:", msg);
        
          setMessages((prevMessages) =>
            prevMessages.map(msglist => {
              if (msglist.id === msg.messageid) {
                return {
                  ...msglist,
                  reaction: msg.reaction // 👍を反映
                };
              }
              return msglist;
            })
          );
          return;
        }
        
        // ✅ 通常のチャットメッセージのみ以下を実行
        if (msg.type !== "postmessage"){
          console.log("postmessage以外は無視");
          return;
        }
        if(!msg.postmessage.Content){
          console.log("msg.content：エラー");
          return;
        }
        if(typeof msg.postmessage.Content !== "string"){
          console.log("typeof msg.content：エラー");
          return;
        }
        if (!msg.postmessage.ID || !msg.postmessage.Content || typeof msg.postmessage.Content !== "string") {
          console.warn("⚠️ 無効なチャットメッセージ:", msg);
          return;
        }        
        console.log("👤：",msg.postmessage.SenderID, userid, msg.postmessage.sendername);


        // ✅ 表示に追加
        const newMessage: Message = {
          id: msg.postmessage.ID,
          sender: msg.postmessage.SenderID,
          sendername: msg.postmessage.SenderName,
          type: msg.postmessage.Content.includes("/uploads/") ? "image" : "text",
          content: msg.postmessage.Content,
          allread: false,
          readcount: 0,
        };
        setMessages((prev) => [...prev, newMessage]);

        // ✅ 既読リクエスト（自分のメッセージは除外）
        if (msg.SenderID !== userid) {
          
        const res = await fetch(`http://localhost:8080/read`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "Authorization": `Bearer ${token}`,
          },
          body: JSON.stringify({ login_id: userid, msg_id: msg.postmessage.ID, room_id: roomId}),
        });
        if (!res.ok) {
          throw new Error("未読一覧取得失敗");
        }

        const data = await res.json();
      }
      } catch (err) {
        console.error("❌ WebSocket受信処理エラー:", err);
      };
    };
    addMessageListener(handleMessage);
    return() => removeMessageListener(handleMessage);


  })

  // メッセージ送信（onClickから呼ばれる）
  const handleSendMessage = async () => {
    const token = localStorage.getItem("token");

    console.log("メッセージ:", messages);
    if (!message.trim()) {
      alert("メッセージを入力してください");
      return;
    }

    try {
      const newMessage = {
        room_id: parseInt(roomId as string, 10),
        sender_id: loggedInUserid,
        content: message.trim(),
      };

      const res = await fetch("http://localhost:8080/message", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${token}`,
        },
        body: JSON.stringify(newMessage),
      });

      if (!res.ok) {
        throw new Error("メッセージ送信失敗");
      }

      const response = await res.json();
      console.log("📨データ：", response);

      // ✅ メンションされたユーザーを抽出（@username を含むかどうか）
      const mentionedUserIds = members
      .filter(member => message.includes(`@${member.username}`))
      .map(member => member.id);

      if (mentionedUserIds.length > 0) {
        await fetch("http://localhost:8080/addMention", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "Authorization": `Bearer ${token}`,
          },
          body: JSON.stringify({
            message_id: parseInt(response.data.ID as string),
            mentioned_target_id: mentionedUserIds,
            room_id: parseInt(roomId as string),
            sender_id: loggedInUserid,
          }),
        });
      }

      setMessage("");
    } catch (err) {
      alert("メッセージ送信エラー");
      console.error("送信エラー:", err);
    }
  };

  // メンションのためのルームメンバー一覧取得
  useEffect(() => {
    const token = localStorage.getItem("token");
    console.log("ユーザーID：",loggedInUserid)
    if (!roomId) return;
    const fetchMembers = async () => {
      const res = await fetch(`http://localhost:8080/getRoomMembers?room_id=${roomId}`,{
        method: "POST",
        headers:{
          "Content-Type": "application/json",
          "Authorization": `Bearer ${token}`,
        },
        body: JSON.stringify({login_id: loggedInUserid}),
      });

      const data = await res.json();
      console.log("メンションデータ：",data.members);
      setMembers(data.members);
    };
    fetchMembers();
  }, [roomId,loggedInUserid]);


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
      console.log("📨ファイルデータ：", response);

      setMessage("");
      
      // ファイル選択をクリア
      setSelectedFile(null);
      if (fileInputRef.current) {
        fileInputRef.current.value = ""; // 実際に選択UIをクリア
      }

    } catch (error) {
      alert("アップロードエラー：" + error);
    }
    console.log("🔍 content:", messages);
  };

  
  // 編集
  const handleEdit = async (id: number) => {
    const token = localStorage.getItem("token");

    console.log("編集：", id);

    if (editText.trim() === "") {
      setIsEditing(false);
      alert("メッセージを入力して下さい");
      return;
    }
    try{
      const res = await fetch(`http://localhost:8080/editMessage?id=${id}`,{
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${token}`,
        },
        body: JSON.stringify({content: editText, room_id: roomId}),
      });

      if(!res.ok) throw new Error("編集失敗");

      //const response = await res.json();

      setMessages((prev) =>
        prev.map((msg) => (msg.id === id ? { ...msg, content: editText } : msg))
      );
      setEditingId(null);
    } catch(error) {
      console.error("保存失敗", error);
      alert("メッセージの更新に失敗しました...")
    }
  }

  // 自分のメッセージを削除
  const handleMyDelete = async (id: number) => {
    const token = localStorage.getItem("token");

    console.log("メッセージ削除📝：", id);
    const confirmed = window.confirm("このメッセージを削除しますか？");
    if (!confirmed) return;

    // messagesaから該当メッセージの削除
    setMessages(messages.filter(msg => msg.id !== id));

    // 削除処理の実装へ
    try{
      const res = await fetch(`http://localhost:8080/deleteMyMessage?id=${id}`, { // id = message.id
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${token}`,
        },
        body: JSON.stringify({ login_id: loggedInUserid,room_id: roomId}),
      });
      if (!res.ok) {
          throw new Error("削除失敗");
        } else {
          alert("メッセージを削除しました");
        }
          // ✅ メッセージを「削除済み表示」に差し替える
          setMessages((prev) =>
            prev.map((msg) =>
              msg.id === id
                ? {
                    ...msg,
                    content: "（このメッセージは削除されました）",
                    type: "text",
                  }
                : msg
            )
          );

          console.log(`🗑️ メッセージ${id}を削除しました`);
      
      } catch (err) {
        alert("削除できませんでした");
        console.error("削除エラー：", err);
      }

  }

  // 送信取消
  const handleDelete = async (id: number) => {
    const token = localStorage.getItem("token");

    const hoveredMessage = messages.find(msg => msg.id === hoveredMessageId);
    console.log("メッセージ送信取消📝：", hoveredMessage);
  
    console.log("送信取消：", id);
    const confirmed = window.confirm("このメッセージを送信取消しますか？");
    if (!confirmed) return;
    
    try{
      const res = await fetch(`http://localhost:8080/deleteMessage?id=${id}`, { // id = message.id
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${token}`,
        },
        body: JSON.stringify({room_id: roomId}),
      });
      if (!res.ok) {
          throw new Error("送信取消失敗");
        } else {
          alert("メッセージを送信取消しました");
        }
          // ✅ メッセージを「削除済み表示」に差し替える
          setMessages((prev) =>
            prev.map((msg) =>
              msg.id === id
                ? {
                    ...msg,
                    content: "（このメッセージは削除されました）",
                    type: "text",
                  }
                : msg
            )
          );

          console.log(`🗑️ メッセージ${id}を送信取消しました`);
      
      } catch (err) {
        alert("送信取消できませんでした");
        console.error("送信取消エラー：", err);
      }

  };
 
  // メンション機能
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setMessage(value);
    setSelectedFile(null); // ファイル入力をリセット

    if (value.endsWith("@")) {
      setShowMentionList(true); // モーダルを表示
    } else {
      setShowMentionList(false); // 非表示
    }
  };



  // メンション相手の表示
  const handleSelectMention = (member: { username: string }) => {
    setMessage((prev) => prev + member.username + " ");
    setShowMentionList(false);
  };

  // トグル
  const toggleUserSelection = (userId: number | undefined) => {
    if (userId === undefined) return;  // safety guard
    setSelectedUsers((prevSelected) =>
      prevSelected.includes(userId)
        ? prevSelected.filter((id) => id !== userId)  // すでに選択されている場合は削除
        : [...prevSelected, userId]  // 選択されていない場合は追加
    );
  };

  //リアクション（message_readsのreactionに追加）
  const handleReact = async (id: number,reaction: string) => {
    console.log("リアクション:", id);
    const token = localStorage.getItem("token");
    const userId = localStorage.getItem("loggedInUserID");
    
    const res = await fetch("http://localhost:8080/addReaction", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${token}`,
      },
      body: JSON.stringify({
        message_id: id,
        user_id: Number(userId),
        room_id: parseInt(roomId as string),
        reaction: reaction,
      }),
    });  

    if (res.ok) {
      // ✅ メッセージ一覧を更新
      setMessages((prev) =>
        prev.map((msg) =>
          msg.id === id ? { ...msg, reaction: reaction } : msg
        )
      );
    }
  };

  // ルーム退出
  const handleLeaveRoom = async () => {
    const token = localStorage.getItem("token");
    const userId = localStorage.getItem("loggedInUserID");
    if (!userId || !roomId) return;
  
    if (!confirm("本当にルームを退出しますか？")) return;
  
    try {
      const res = await fetch("http://localhost:8080/leaveRoom", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${token}`,
        },
        body: JSON.stringify({
          room_id: parseInt(roomId as string),
          user_id: parseInt(userId),
        }),
      });
  
      if (!res.ok) throw new Error("退出失敗");
  
      alert("ルームから退出しました");
      router.push("/roomSelect"); // 戻るなどのリダイレクト
    } catch (err) {
      console.error("退出エラー:", err);
      alert("退出に失敗しました");
    }
  };

  // メンバー追加のためのユーザー一覧取得
  useEffect(() => {
    const fetchNotMembers = async () => {
      const token = localStorage.getItem("token");
      const userId = localStorage.getItem("loggedInUserID");
      const i_userId = userId !== null ? parseInt(userId, 10) : null;
      const membersArray = Array.isArray(members) ? members : Object.values(members);

      console.log("送るmembers：", members, Array.isArray(members)); 

      if (!roomId) return;
        const res = await fetch(`http://localhost:8080/usersNotInRoom?room_id=${roomId}`,{
          method: "POST",
          headers:{
            "Content-Type": "application/json",
            "Authorization": `Bearer ${token}`,
          },
          body: JSON.stringify({login_id: i_userId, members: membersArray}),
        });

        if(!res.ok){
          console.log("他のユーザーを取得できません");
        }

        const data = await res.json();
        console.log("メンバー以外のユーザー：", data.members)
        setNotMembers(data.members);
  }
  fetchNotMembers();
}, [roomId]);

  // メンバー追加
  const handleAddMember = async () => {
    const token = localStorage.getItem("token");
    const userId = localStorage.getItem("loggedInUserID");

    if (!userId || !roomId) return;
    try {
      const res = await fetch("http://localhost:8080/addMember", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${token}`,
        },
        body: JSON.stringify({ 
          room_id: parseInt(roomId as string),
          user_ids: selectedUsers,
        }),
      });
  
      if (!res.ok) throw new Error("退出失敗");
      closePersonalModal();
      alert("メンバーを追加しました");
      window.location.href = location.pathname;
    } catch (err) {
      console.error("退出エラー:", err);
      alert("メンバー追加に失敗しました");
    }

  };

  // ルーム作成モーダル
  const openModal = () => {
    setSelectedUsers([]);
    setIsModalOpen(true);
  }
  const closeModal = () => setIsModalOpen(false);

  const closePersonalModal = () => setIsPersonalModalOpen(false);
  
  


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
        {/* ルーム名 */}
        <div style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: "20px"
        }}>
        <h2 style={{ color: "#388e3c", margin: 0 }}>
          ルーム：{groupName ? groupName : "ルーム名がありません"}
        </h2>

        {/* ボタン群（右寄せ） */}
        <div style={{ display: "flex", gap: "10px" }}>
          <button
            onClick={handleLeaveRoom}
            style={{
              backgroundColor: "#e8f5e9",
              color: "#333",
              border: "1px solid #ccc",
              borderRadius: "12px",
              padding: "6px 12px",
              fontSize: "12px",
              fontWeight: "bold",
              cursor: "pointer"
            }}
          >
            退出する
          </button>
          {isGroup == 1 &&(
          <button
            onClick={openModal}
            style={{
              backgroundColor: "#e8f5e9",
              color: "#333",
              border: "1px solid #ccc",
              borderRadius: "12px",
              padding: "6px 12px",
              fontSize: "12px",
              fontWeight: "bold",
              cursor: "pointer"
            }}
          >+ メンバー追加</button>)}
            {isModalOpen && (
              <div style={{ fontSize: "18px",position: "fixed", top: "20%", left: "50%", transform: "translate(-50%, -20%)", backgroundColor: "#fff", padding: "20px", borderRadius: "10px", boxShadow: "0 4px 8px rgba(0,0,0,0.2)", width: "40%",
                maxWidth: "400px", zIndex: 1000, }}>
                <h3>グループ作成</h3>
                {notMembers.map((notmembers) => (
                  <div key={notmembers.id } style={{ display: "flex", alignItems: "center", justifyContent: "flex-start", marginBottom: "8px" }}>
                    <input type="checkbox" style={{ marginRight: "20px", marginLeft:"50px" }} checked={selectedUsers.includes(notmembers.id)} onChange={() => toggleUserSelection(notmembers.id)} />
                    {notmembers.username}
                  </div>
                ))}
                  <button onClick={handleAddMember} style={{ padding: "8px 16px", margin: "10px", backgroundColor: "#388e3c", color: "#fff", borderRadius: "20px" }}>追加</button>
                  <button onClick={closeModal} style={{ padding: "8px 16px", margin: "10px", backgroundColor: "#388e3c", color: "#fff", borderRadius: "20px" }}>キャンセル</button>
                </div>
                    )}
                    </div>  
                </div>
        
        <div style={{ maxHeight: "500px", overflowY: "scroll", marginBottom: "15px" }}>
          {messages.length >= 0 ? (
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
                    }, 700); // 700ms待って表示
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
                      
                    {/* 本文 or 画像 */}
                    {(msg.content).startsWith("http") &&
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
                        {msg.content}
                      </div>
                    )}
                  </>
                )}
                    {/* 既読 */}
                    {isMyMessage && (
                      <div
                        style={{
                          fontSize: "11px",
                          color: "green",
                          position: "absolute",
                          bottom: "1px",
                          right: "10px",
                        }}
                      >
                        {msg.allread ? "全員既読" : `既読 ${msg.readcount-1}`}
                      </div>
                    )}
                  </div>
                {/* 吹き出しの右にリアクション */}
                {msg.reaction && (
                  <div
                    style={{
                      display: "flex",
                      gap: "4px",
                      fontSize: "20px",
                      marginLeft: "4px",
                      userSelect: "none",
                    }}
                  >
                    {msg.reaction
                      .split(",")                       // カンマで分割
                      .filter((r) => r.trim() !== "")   // 空文字を除外
                      .map((emoji, i) => (
                        <span key={i}>{emoji}</span>    // 一つずつ表示
                      ))}
                  </div>
                )}
                  {/* ホバーメニュー */}
                  {hoveredMessageId === msg.id && (
                    <div
                      style={{
                        position: "absolute",
                        bottom: "-26px",
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
                            fontSize: "10px",
                            cursor: "pointer",
                          }}
                          onClick={() => {
                            setEditingId(msg.id);
                            setEditText(msg.content);
                          }}
                          >編集</span>
                          <span
                          style={{
                            fontSize: "10px",
                          }}
                          onClick={() => handleMyDelete(msg.id)}>削除</span>
                          <span
                          style={{
                            fontSize: "10px",
                          }}
                          onClick={() => handleDelete(msg.id)}>送信取消</span>
                        </>
                      ) : (
                        <span
                          style={{ fontSize: "13px", cursor: "pointer" }}
                          onClick={() => handleReact(msg.id, "👍")}
                        >
                          👍
                        </span>
                        
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
            <button onClick={() => setShowEmojiPicker(prev => !prev)}> 😊 </button>
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
            onChange={ handleInputChange }
            style={{
              flex: 1,
              margin: "0 10px",
              padding: "16px",
              borderRadius: "30px",
              border: "2px solid #ccc"
            }}
          />

          {/* メンション機能 */}
          {showMentionList && (
            <div style={{ 
              position: "absolute", 
              bottom: "60px", 
              left: "300px", 
              backgroundColor: "#fffde7", 
              borderRadius: "20px",
              zIndex: 200 
              }}>
              {members.map((member) => (
                <div
                  key={member.id}
                  style={{ padding: "5px 10px", cursor: "pointer", color: "blue" }}
                  onClick={() => handleSelectMention(member)}
                >
                  @{member.username}
                </div>
              ))}
            </div>
            )}

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
