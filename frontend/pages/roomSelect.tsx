import Head from "next/head";
import { useRouter } from "next/navigation";
import { useState, useEffect } from "react";
import { connectWebSocket, addMessageListener, removeMessageListener } from "../utils/websocket";


interface User {
  id: number;
  username: string;
}
interface Room {
  id: number;
  room_name: string;
  unread_count: number;
  unread_mention_count: number; // ← これが正しく認識されていればOK
  is_group: number;
}

export default function RoomSelect() {
  const [loggedInUser, setLoggedInUser] = useState<string | null>(null);
  const [loggedInUserID, setLoggedInUserID] = useState<number | null>(null);
  const [users, setUsers] = useState<User[]>([]);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isPersonalModalOpen, setIsPersonalModalOpen] = useState(false);
  const [groupName, setGroupName] = useState("");
  const [personals, setPersonals] = useState<Room[]>([]);
  const [rooms, setRooms] = useState<Room[]>([]);
  const [selectedRoom, setSelectedRoom] = useState<Room | null>(null);
  const [selectedUsers, setSelectedUsers] = useState<number[]>([]);
  const router = useRouter();

  // ログインしているか確認
  useEffect(() => {
    console.log(localStorage);
    const loggedInUsername = localStorage.getItem("loggedInUser");
    const loggedInUserIDStr = localStorage.getItem("loggedInUserID");
    if (loggedInUsername && loggedInUserIDStr) {
      setLoggedInUser(loggedInUsername);
      setLoggedInUserID(parseInt(loggedInUserIDStr, 10));
    } else {
      alert("ログインが必要です");
      router.push("/top");
    }
  }, []);

  // 所属している個別ルーム一覧の取得
  useEffect(() => {
    const fetchPersonalRooms = async () => {
      try {
        const token = localStorage.getItem("token");
        if (!token) {
          alert("ログインされていません");
          router.push("/top");
          return;
        }

        const loggedIDStr = localStorage.getItem("loggedInUserID");
    
        const res = await fetch("http://localhost:8080/PersonalRoomSelect", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "Authorization": `Bearer ${token}`,
          },
          body: JSON.stringify({ login_id: loggedIDStr !== null ? parseInt(loggedIDStr) : null })
        });

        if (!res.ok) {
          throw new Error("ルーム一覧取得失敗");
        }

        const data = await res.json();

        if (Array.isArray(data)) {
          console.log("🟣個別ルーム取得:",data);
          setPersonals(data);
        }
      } catch (err) {
        console.error("ルーム一覧取得エラー：", err);
      }
    };
    fetchPersonalRooms();
  }, []);

  // 所属しているグループルーム一覧の取得
  useEffect(() => {
    const fetchRooms = async () => {
      try {
        const token = localStorage.getItem("token");
        if (!token) {
          alert("ログインされていません");
          router.push("/top");
          return;
        }

        const loggedIDStr = localStorage.getItem("loggedInUserID");
    
        const res = await fetch("http://localhost:8080/groupRoomSelect", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "Authorization": `Bearer ${token}`,
          },
          body: JSON.stringify({ login_id: loggedIDStr !== null ? parseInt(loggedIDStr) : null })
        });

        if (!res.ok) {
          throw new Error("ルーム一覧取得失敗");
        }

        const data = await res.json();
        console.log("🟣グループルーム取得：",data)
        if (Array.isArray(data)) {
          setRooms(data);
        }
      } catch (err) {
        console.error("ルーム一覧取得エラー：", err);
      }
    };
    fetchRooms();
  }, []);

  // ルーム作成のときのユーザー一覧(モーダル)の取得
  useEffect(() => {
    const fetchUsers = async () => {
      try {
        const token = localStorage.getItem("token");
        if (!token) {
          alert("ログインされていません");
          router.push("/top");
          return;
        }

        const loggedIDStr = localStorage.getItem("loggedInUserID");

        const res = await fetch("http://localhost:8080/roomSelect", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "Authorization": `Bearer ${token}`,
          },
          body: JSON.stringify({ login_id: loggedIDStr !== null ? parseInt(loggedIDStr) : null })
        });

        if (!res.ok) {
          throw new Error("ユーザー一覧取得失敗");
        }

        const data = await res.json();
        if (Array.isArray(data)) {
          setUsers(data);
        }
      } catch (err) {
        console.error("ユーザー一覧取得エラー：", err);
      }
    };
    fetchUsers();
  }, []);

  //個別ルーム作成
  const handleCreatePersonal = async () => {
    const token = localStorage.getItem("token");

    if (selectedUsers.length != 1) {
      alert("メンバーを選択してください");
      return;
    }
    try {
      const res = await fetch("http://localhost:8080/createRooms", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${token}`,
        },
          body: JSON.stringify({ room_name: groupName, user_ids: selectedUsers, login_id: loggedInUserID})
      });
      if (!res.ok) {
        const errorMessage = await res.text();  // エラーメッセージを取得
        throw new Error(`エラー: ${errorMessage}`);
      } 
      closePersonalModal();
      alert("ルームを作成しました");
    } catch (err) {
      if (err instanceof Error) {
        console.error("エラー:", err.message);
        alert(`${err.message}`);
      } else {
        console.error("未知のエラー:", err);
        alert(`サーバーエラー: ${String(err)}`);
      }
    }
  };

  //グループルーム作成
  const handleCreateGroup = async () => {
    if (!groupName || selectedUsers.length < 2) {
      alert("グループ名とメンバーを選択してください");
      return;
    }
    const token = localStorage.getItem("token");

    try {
      const res = await fetch("http://localhost:8080/createGroup", {
        method: "POST",
        headers: { 
          "Content-Type": "application/json",
          "Authorization": `Bearer ${token}`
        },
        body: JSON.stringify({ room_name: groupName, user_ids: selectedUsers, login_id: loggedInUserID})
      });
      if (!res.ok) {
        const errorMessage = await res.text();  // エラーメッセージを取得
        throw new Error(`エラー: ${errorMessage}`);
      } 
      closeModal();
      alert("グループを作成しました");
    } catch (err) {
      if (err instanceof Error) {
        console.error("エラー:", err.message);
        alert(`${err.message}`);
      } else {
        console.error("未知のエラー:", err);
        alert(`サーバーエラー: ${String(err)}`);
      }
    }
  };

  // ルームを選択してルームへ
  const handleSelectRoom = async (room: Room) => {
    try {
      console.log("チャットルームへ");
      const token = localStorage.getItem("token");
      if (!token) {
        alert("ログインされていません");
        router.push("/top");
        return;
      }
      const userIDStr = localStorage.getItem("loggedInUserID");
      if (!userIDStr) {
        alert("ログインされていません");
        router.push("/top");
        return;
      }
      if (!room.id || room.id === 0) {
        alert("ルームIDが無効です");
        return;
    }

      console.log(room);
      setSelectedRoom(room);

      localStorage.setItem("token", token);
      localStorage.setItem("roomName", room.room_name);
      localStorage.setItem("is_group", String(room.is_group));
      router.push(`/${room.id}`);
    } catch (err) {
      console.error("ルーム遷移エラー：", err);
    }
  };

  // メッセージ受け取り
  useEffect(() => {
    connectWebSocket();

    const handleMessage = (msg: any) => {
      const loginUserID = localStorage.getItem("loggedInUserID");
      const i_loginUserID = loginUserID ? parseInt(loginUserID, 10) : null;

      console.log("handleMessage:",msg);

      if (msg.type === "unreadmessage") {
        console.log("🔔 未読通知を受信:", msg.userId);
        interface SendMessages {
          user_id: number;
          room_id: number;
          unread_count: number;
        }

        // SendMessagesをMapに変換して高速アクセス
        const sendMap = new Map<number, SendMessages>();
        for (const sm of msg.unReadMessage) {

          if(i_loginUserID === sm.user_id){
            sendMap.set(sm.room_id, sm);
          }
        }

        // personalsを上書きして新しい未読配列を返す
        setPersonals((prevPersonals) =>
          prevPersonals.map(personallist => {
            console.log("Personal.mapスタート ");
            const readInfo = sendMap.get(personallist.id);
            if (readInfo) {
              console.log("readInfo:", personallist.id, " > ", personallist.room_name, " > ", personallist.unread_count);
              return {
                ...personallist,
                unread_count: readInfo.unread_count
              };
            }
            return personallist;
          })
        );
        
        // rooms(group)を上書きして新しい未読配列を返す
        setRooms((prevRooms) =>
          prevRooms.map(roomlist => {
            console.log("GroupRoom.mapスタート ");
            const readInfo = sendMap.get(roomlist.id);
            if (readInfo) {
              console.log("readInfo:", roomlist.id, " > ", roomlist.room_name, " > ", roomlist.unread_count);
              return {
                ...roomlist,
                unread_count: readInfo.unread_count
              };
            }
            return roomlist;
          })
        );
      }
      
      if (msg.type === "mention") {
        console.log("🔔 メンション通知受信:", msg);
        interface SendMessages {
          user_id: number;
          room_id: number;
          unread_mention_count: number;
        }

        const mentionMap = new Map<number, SendMessages>();
        for (const sm of msg.Mention) {

          if(i_loginUserID === sm.user_id){
            mentionMap.set(sm.room_id, sm);
          }
        }

        setRooms((prevRooms) =>
          prevRooms.map(roomlist => {
            console.log("GroupRoom.mapスタート ");
            const readInfo = mentionMap.get(roomlist.id);
            if (readInfo) {
              console.log("readInfo:", roomlist.id, " > ", roomlist.room_name, " > ", roomlist.unread_mention_count);
              return {
                ...roomlist,
                unread_mention_count: readInfo.unread_mention_count
              };
            }
            return roomlist;
          })
        );

      }

      if(msg.type === "createroom"){
        console.log("🔔 ルーム作成通知受信:", msg);

        // 自分に、該当ルームか確認
        console.log("i_loginUserID:", i_loginUserID, "msg.memberlist",msg.memberlist);
        var roomname = ""
        for (let membercount = 0; membercount < msg.memberlist.length; membercount++){
          if(msg.memberlist[membercount].user_id === i_loginUserID){
            roomname = msg.memberlist[membercount].group_name;
          }
        }
        if (roomname === ""){
          console.log("作成されたルームはログインユーザーとは無関係です");
          return;
        }

        const newRoom: Room = {
          id: msg.room_id,
          room_name: roomname,
          unread_count: 0,
          unread_mention_count: 0,
          is_group: msg.is_group,
        };

        if (msg.is_group === 0){
          setPersonals((prev) => {
            const exists = prev.some((personal) => personal.id === msg.room_id);
            if (exists) return prev;
          
            return [...prev, newRoom];
          });
        }

        if (msg.is_group == 1){
          setRooms((prev) => {
            const exists = prev.some((room) => room.id === msg.room_id);
            if (exists) return prev;
          
            return [...prev, newRoom];
          });
        }
      }

      if (msg.type === "leaveroom"){
        console.log("🔔 退出通知受信:", msg);

        console.log("i_loginUserID:", i_loginUserID, "msg.userids",msg.userids);
        if (!(msg.userids.includes(i_loginUserID))){
          console.log("退出するルームがありません");
          return
        }
        setPersonals((prevPersonals) =>
          prevPersonals.filter((personal) => personal.id !== msg.room_id)
        );
        
      }

      if (msg.type === "addmembers"){
        console.log("🔔 メンバー追加通知受信:", msg);
        alert("グループに招待されました");
        window.location.href = location.pathname;

        console.log("room_id:", msg.room_id, "msg.userids",msg.userids);
        if (!(msg.userids.includes(i_loginUserID))){
          console.log("退出するルームがありません");
          return
        }
        
      }



    };

    addMessageListener(handleMessage);
    return() => removeMessageListener(handleMessage);
  }, []);

  // personalsの更新
  useEffect(() => {
    console.log("✅更新された personals:", personals);
  }, [personals]);

  // ユーザー選択をトグルする関数
  const toggleUserSelection = (userId: number) => {
    setSelectedUsers((prevSelected) =>
      prevSelected.includes(userId)
        ? prevSelected.filter((id) => id !== userId)  // すでに選択されている場合は削除
        : [...prevSelected, userId]  // 選択されていない場合は追加
    );
  };



  // ルーム作成モーダル
  const openModal = () => {
    setSelectedUsers([]);
    setIsModalOpen(true);
  }
  const closeModal = () => setIsModalOpen(false);

  const openPersonalModal = () => {
    setSelectedUsers([]);
    setIsPersonalModalOpen(true);
  }
  const closePersonalModal = () => setIsPersonalModalOpen(false);

  //ログアウト
  const handleLogout = () => {
    localStorage.removeItem("loggedInUser");
    localStorage.removeItem("token");
    alert("ログアウトしました");
    router.push("/top");
  };

  return (
    <>
      <Head>
        <title>チャットルーム選択</title>
      </Head>
      <div
        style={{
          background: "linear-gradient(180deg, #e8f5e9, #fffde7)",
          minHeight: "100vh",
          height: "100vh",            // 明示的に高さを指定
          overflow: "hidden",         // スクロールを抑制
          display: "flex",
          flexDirection: "column",
          justifyContent: "center",
          alignItems: "center"
      }}>
        <div style={{
          background: "white",
          padding: "40px",
          borderRadius: "30px",
          boxShadow: "0px 8px 16px rgba(0,0,0,0.1)",
          textAlign: "center",
          width: "90%",
          maxWidth: "1000px"
        }}>
          <h2 style={{ color: "#388e3c", fontWeight: "bold", marginBottom: "15px" }}>チャットルーム選択</h2>
          <p style={{ color: "#555", marginBottom: "25px", fontSize: "16px" }}>ログイン中: {loggedInUser}</p>
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", gap: "40px", marginBottom: "10px" }}>
            {/* 個別チャット */}
            <div style={{
              flex: 5,
              height: "450px",      // 高さを固定
              overflowY: "scroll",    // スクロールを有効にする
              alignItems: "flex-start"}}>
              <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "10px" }}>
                <div style={{ flex: 1 }}></div>
                  <h3 style={{ color: "#388e3c", marginBottom: "0px", textAlign: "center",flex: 1 }}>個別ルーム</h3>
                  <button onClick={openPersonalModal} style={{ padding: "8px 16px", margin: "10px", backgroundColor: "#388e3c", color: "#fff", borderRadius: "20px" }}>＋ユーザー追加</button>
                  {isPersonalModalOpen && (
                      <div style={{ fontSize: "18px",position: "fixed", top: "20%", left: "50%", transform: "translate(-50%, -20%)", backgroundColor: "#fff", padding: "20px", borderRadius: "10px", boxShadow: "0 4px 8px rgba(0,0,0,0.2)", width: "40%",
                        maxWidth: "400px" }}>
                        <h3>ルーム作成</h3>
                        {users.map((user) => (
                          <div key={user.id } style={{ display: "flex", alignItems: "center", justifyContent: "flex-start", marginBottom: "8px" }}>
                            <input type="checkbox" style={{ marginRight: "20px", marginLeft:"50px" }} checked={selectedUsers.includes(user.id)} onChange={() => toggleUserSelection(user.id)} />
                            {user.username}
                          </div>
                        ))}
                        <button onClick={handleCreatePersonal} style={{ padding: "8px 16px", margin: "10px", backgroundColor: "#388e3c", color: "#fff", borderRadius: "20px" }}>作成</button>
                        <button onClick={closePersonalModal} style={{ padding: "8px 16px", margin: "10px", backgroundColor: "#388e3c", color: "#fff", borderRadius: "20px" }}>キャンセル</button>
                      </div>
                  )}
              </div>
              {personals.map((personal) => (
                <div
                  key={personal.id}
                  onClick={() => handleSelectRoom(personal)}
                  style={{
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "flex-start",
                    padding: "15px",
                    margin: "10px auto",
                    width: "100%",
                    maxWidth: "800px",
                    cursor: "pointer",
                    backgroundColor: selectedUser?.id === personal.id ? "#c8e6c9" : "#ffffff",
                    borderRadius: "30px",
                    boxShadow: "0 5px 4px rgba(0, 0, 0, 0.1)",
                    transition: "all 0.3s",
                  }}
                >
                <div style={{ backgroundColor: "#81c784", width: "10px", height: "10px", borderRadius: "50%", marginRight: "15px" }}></div>
                  <span style={{ color: "#333", fontSize: "18px", textAlign: "left" }}>{personal.room_name}</span>
                    {/* 未読通知：個別チャット */}
                    {personal.unread_count != 0 && (
                      <div style={{   
                        backgroundColor: '#d02f2f',
                        color: 'white',
                        borderRadius: '9999px',
                        padding: '4px 8px',
                        fontSize: '12px',
                        fontWeight: 'bold',
                        marginLeft: 'auto',
                        marginRight: '10px',
                        boxShadow: '0 2px 4px rgba(0, 0, 0, 0.2)'}}>{personal.unread_count}</div>
                    )}
                </div>
              ))}
            </div>
            {/* グループチャット */}
            <div style={{
              flex: 5,
              height: "460px",      // 高さを固定
              overflowY: "scroll",    // スクロールを有効にする
              alignItems: "flex-start"  // 上揃えに変更
              }}>
              <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "10px" }}>
                <div style={{ flex: 1 }}></div>
                  <h3 style={{ color: "#388e3c", marginBottom: "0px", textAlign: "center",flex: 1 }}>グループルーム</h3>
                  <button onClick={openModal} style={{ padding: "8px 16px", margin: "10px", backgroundColor: "#388e3c", color: "#fff", borderRadius: "20px" }}>＋グループ作成</button>
                  {isModalOpen && (
                    <div style={{ fontSize: "18px",position: "fixed", top: "20%", left: "50%", transform: "translate(-50%, -20%)", backgroundColor: "#fff", padding: "20px", borderRadius: "10px", boxShadow: "0 4px 8px rgba(0,0,0,0.2)", width: "40%",
                      maxWidth: "400px" }}>
                      <h3>グループ作成</h3>
                      <input type="text" placeholder="グループ名" value={groupName} onChange={(e) => setGroupName(e.target.value)} />
                      {users.map((user) => (
                        <div key={user.id } style={{ display: "flex", alignItems: "center", justifyContent: "flex-start", marginBottom: "8px" }}>
                          <input type="checkbox" style={{ marginRight: "20px", marginLeft:"50px" }} checked={selectedUsers.includes(user.id)} onChange={() => toggleUserSelection(user.id)} />
                          {user.username}
                        </div>
                      ))}
                      <button onClick={handleCreateGroup} style={{ padding: "8px 16px", margin: "10px", backgroundColor: "#388e3c", color: "#fff", borderRadius: "20px" }}>作成</button>
                      <button onClick={closeModal} style={{ padding: "8px 16px", margin: "10px", backgroundColor: "#388e3c", color: "#fff", borderRadius: "20px" }}>キャンセル</button>
                    </div>
                  )}
              </div>
              {rooms.map((room) => (
                <div
                  key={room.id}
                  onClick={() => handleSelectRoom(room)}
                  style={{
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "flex-start",
                    padding: "15px",
                    margin: "10px auto",
                    width: "100%",
                    maxWidth: "800px",
                    cursor: "pointer",
                    backgroundColor: selectedUser?.id === room.id ? "#c8e6c9" : "#ffffff",
                    borderRadius: "30px",
                    boxShadow: "0 5px 4px rgba(0, 0, 0, 0.1)",
                    transition: "all 0.3s",
                  }}
                >
                  <div style={{ backgroundColor: "#81c784", width: "10px", height: "10px", borderRadius: "50%", marginRight: "15px" }}></div>
                  <span style={{ color: "#333", fontSize: "18px", textAlign: "left" }}>{room.room_name}</span>
                  {room.unread_mention_count != 0 && (
                    <div style={{   
                      backgroundColor: '#426AB3',
                      color: 'white',
                      borderRadius: '9999px',
                      padding: '4px 8px',
                      fontSize: '9px',
                      fontWeight: 'bold',
                      marginLeft: 'auto',
                      marginRight: '10px',
                      boxShadow: '0 2px 4px rgba(0, 0, 0, 0.2)'}}>@ メンションされました</div>
                  )}
                  {room.unread_count != 0 && (
                      <div style={{   
                        backgroundColor: '#d02f2f',
                        color: 'white',
                        borderRadius: '9999px',
                        padding: '4px 8px',
                        fontSize: '12px',
                        fontWeight: 'bold',
                        marginLeft: 'auto',
                        marginRight: '10px',
                        boxShadow: '0 2px 4px rgba(0, 0, 0, 0.2)'}}>{room.unread_count}</div>
                    )}
                </div>
              ))}
            </div>
          </div>
          <footer style={{ marginTop: "30px", textAlign: "center" }}>
            <button onClick={handleLogout} style={{
              backgroundColor: "#388e3c",
              color: "#fff",
              padding: "10px 25px",
              borderRadius: "20px",
              boxShadow: "0 2px 4px rgba(0,0,0,0.1)",
              transition: "all 0.3s"
            }}>ログアウト</button>
          </footer>
        </div>
      </div>
    </>
  );
}
