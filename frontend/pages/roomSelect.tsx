import Head from "next/head";
import { useRouter } from "next/navigation";
import { Geist, Geist_Mono } from "next/font/google";
import { useState, useEffect } from "react";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

interface User {
  id: number;
  username: string;
}
interface Room {
  id: number;
  room_name: string;
}
interface Member {
  room_id: number;
  room_name: string;
  is_group: number;
  user_id: number | null;
}

export default function RoomSelect() {
  const [loggedInUser, setLoggedInUser] = useState<string | null>(null);
  const [loggedInUserID, setLoggedInUserID] = useState<number | null>(null);
  const [users, setUsers] = useState<User[]>([]);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [members, setMembers] = useState<Member[]>([]);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isPersonalModalOpen, setIsPersonalModalOpen] = useState(false);
  const [groupName, setGroupName] = useState("");
  const [personals, setPersonals] = useState<Room[]>([]);
  const [rooms, setRooms] = useState<Room[]>([]);
  const [selectedRoom, setSelectedRoom] = useState<Room | null>(null);
  const [selectedUsers, setSelectedUsers] = useState<number[]>([]);
  const router = useRouter();

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

  //グループルーム作成
  const handleCreateGroup = async () => {
    if (!groupName || selectedUsers.length < 2) {
      alert("グループ名とメンバーを選択してください");
      return;
    }

    console.log("🟢",selectedUsers,loggedInUserID);
    try {
      const res = await fetch("http://localhost:8080/createGroup", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ room_name: groupName, user_ids: selectedUsers, login_id: loggedInUserID})
      });
      if (!res.ok) {
        const errorMessage = await res.text();  // エラーメッセージを取得
        throw new Error(`エラー: ${errorMessage}`);
      } 
      closeModal();
      alert("グループを作成しました");
      window.location.href = location.pathname;
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

  //個別ルーム
  const handleCreatePersonal = async () => {
    console.log("🟢",selectedUsers,loggedInUserID);
    if (selectedUsers.length != 1) {
      alert("メンバーを選択してください");
      return;
    }
    try {
      const res = await fetch("http://localhost:8080/createRooms", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ room_name: groupName, user_ids: selectedUsers, login_id: loggedInUserID})
      });
      if (!res.ok) {
        const errorMessage = await res.text();  // エラーメッセージを取得
        throw new Error(`エラー: ${errorMessage}`);
      } 
      closePersonalModal();
      alert("ルームを作成しました");
      window.location.href = location.pathname;
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



  // ユーザー選択をトグルする関数
  const toggleUserSelection = (userId: number) => {
    setSelectedUsers((prevSelected) =>
      prevSelected.includes(userId)
        ? prevSelected.filter((id) => id !== userId)  // すでに選択されている場合は削除
        : [...prevSelected, userId]  // 選択されていない場合は追加
    );
  };

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

  // ユーザー一覧の取得
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
        console.log("BBBBBB",loggedIDStr);

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

        const loggedUsername = localStorage.getItem("loggedInUser");
        const loggedIDStr = localStorage.getItem("loggedInUserID");
    
        // var aaa = loggedIDStr !== null ? parseInt(loggedIDStr) : null 

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
        console.log("🟣Personal：",data)
        if (Array.isArray(data)) {
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

        const loggedUsername = localStorage.getItem("loggedInUser");
        const loggedIDStr = localStorage.getItem("loggedInUserID");
    
        var aaa = loggedIDStr !== null ? parseInt(loggedIDStr) : null 

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
        console.log("🟣",data)
        if (Array.isArray(data)) {
          setRooms(data);
        }
      } catch (err) {
        console.error("ルーム一覧取得エラー：", err);
      }
    };
    fetchRooms();
  }, []);


  
  // ユーザーを選択して個別ルームへ (使ってない)
  const handleSelectUser = async (user: User) => {
    try {
      const userIDStr = localStorage.getItem("loggedInUserID");
      if (!userIDStr) {
        alert("ログインされていません");
        router.push("/top");
        return;
      }
      const userID = parseInt(userIDStr, 10);
      const selectedUserID = user.id;
      setSelectedUser(user);

      const token = localStorage.getItem("token");
      const res = await fetch(`http://localhost:8080/createRooms`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${token}`,
        },
        body: JSON.stringify({ user1: userID, user2: selectedUserID}),
      });
      

      if (!res.ok) {
        throw new Error("チャットルーム作成失敗");
      }
      const data = await res.json();
      if (data && data.roomId) {
        localStorage.setItem("token", token ? token : "");
        localStorage.setItem("roomName", data.roomId);
  
      } else {
        alert("ルームIDが取得できませんでした");
      }
    } catch (err) {
      console.error("ルーム遷移エラー：", err);
    }
  };

  // ルームを選択してルームへ
  const handleSelectRoom = async (room: Room) => {
    try {
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
      router.push(`/${room.id}`);
    } catch (err) {
      console.error("ルーム遷移エラー：", err);
    }
  };

  // ルーム作成時にもリアルタイム反映をさせた
  useEffect(() => {
    // // WebSocket接続の処理
    // const socket = createWebSocket((message: any) => {
    //   console.log("受信したメッセージ:", message);
    //   const newRoom = JSON.parse(message);

    //   // 新しいルームがグループルームの場合
    //   if (newRoom.is_group === 1) {
    //     setRooms((prevRooms) => [...prevRooms, newRoom]); // グループルームを追加
    //   } else {
    //     setPersonals((prevPersonals) => [...prevPersonals, newRoom]); // 個別ルームを追加
    //   }
    // });

    // socket.onopen = () => {
    //   console.log("WebSocket接続成功！");
    // };

    // return () => {
    //   socket.close();
    // };
  }, []); // 空の依存配列なので、一度だけ実行される


  const handleLogout = () => {
    localStorage.removeItem("loggedInUser");
    localStorage.removeItem("token");
    alert("ログアウトしました");
    router.push("/top");
    //socket.close();
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
