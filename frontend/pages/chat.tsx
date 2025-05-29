// "use client";


// // 使ってないファイル



// import { useState, useEffect } from "react";
// import { useRouter } from "next/navigation";
// import styles from "@/styles/Home.module.css";
// import axios from "axios";

// interface User {
//   id: number;
//   username: string;
// }

// interface Message {
//   id: number;
//   sender: string;
//   content: string;
// }

// export default function ChatApp() {
//   //const [users, setUsers] = useState<User[]>([]);
//   //const [selectedUser, setSelectedUser] = useState<User | null>(null);
//   const [messages, setMessages] = useState<Message[]>([]);
//   const [message, setMessage] = useState("");
//   const [loggedInUser, setLoggedInUser] = useState<string | null>(null);
//   const router = useRouter();

//   // ログイン中のユーザー名を取得
//   useEffect(() => {
//     const loggedInUsername = localStorage.getItem("loggedInUser");
//     console.log("ログインユーザー", loggedInUsername);
//     if (loggedInUsername) {
//       setLoggedInUser(loggedInUsername);
//     }
//   }, []);

//   // ユーザー一覧を取得
//   useEffect(() => {
//     localStorage.getItem("token");        
//     const fetchLoggedInUser = async () => {
//       try {
//         // セッションストレージからトークンを取得
//         const token = localStorage.getItem("token");
//         if (token) {
//           const res = await fetch("http://localhost:8080/chat", {
//             headers: {
//               "Authorization": `Bearer ${token}`,
//             },
//           });

//           // ユーザー名をセット
//         } else {
//           console.error("トークンがありません");
//           alert("ログインされていません");
//           router.push("/top");
//         }
//       } catch (err) {
//         console.error("ログインユーザー取得エラー:", err);
//       }
//     };
//     fetchLoggedInUser();
//   }, []);

//    // メッセージを表示
//   // const fetchMessages = async () => {
//   //   try {
//   //     const res = await axios.get(`http://localhost:8080/chat?user=${selectedUser?.id}`, {
//   //       withCredentials: true,
//   //     });
//   //     setMessages(res.data);
//   //   } catch (err) {
//   //     console.error("メッセージ取得エラー:", err);
//   //   }
//   // };

//   // メッセージ送信
//   const handleSendMessage = async () => {
//     if (!message.trim()) {
//       alert("メッセージを入力してください");
//       return;
//     }

//     // メッセージ送信処理、フロントエンドで一時的に保存
//     try {
//       const newMessage: Message = {
//         id: messages.length + 1,
//         sender: loggedInUser || "未ログイン",
//         content: message.trim(),
//       };
//       setMessages([...messages, newMessage]);
//       setMessage("");
//       console.log("メッセージ送信成功");
//     } catch (err) {
//       console.error("メッセージ送信エラー:", err);
//       alert("メッセージ送信失敗");
//     }
//   };

//   // トークンを削除する関数
//   const handleLogout = () => {
//     localStorage.removeItem("loggedInUser");
//     localStorage.removeItem("token");
//     const token = sessionStorage.getItem("token");
//     if (!token) {
//       console.log("トークンが正常に削除されました");
//     } else {
//       console.log("トークン削除に失敗しました");
//     }
//     alert("ログアウトしました");
//     console.log("トークン:", token);
//     router.push("/top");
//   };  

//   return (
//     <div style={{ display: "flex", height: "100vh" }}>
//       {/* 左カラム：ユーザー一覧 */}
//       {/* <div style={{ width: "20%", borderRight: "1px solid #ccc", padding: "10px" }}>
//         <h3>ユーザー一覧</h3>
//         <p>ログイン中: {loggedInUser || "未ログイン"}</p>
//         {users.length > 0 ? (
//           users.map((user) => (
//             <div
//               key={user.id}
//               onClick={() => {
//                 setSelectedUser(user);
//                 //fetchMessages();
//               }}
//               style={{
//                 padding: "8px",
//                 cursor: "pointer",
//                 backgroundColor: selectedUser?.id === user.id ? "#e0e0e0" : "white",
//               }}
//             >
//               {user.username}
//             </div>
//           ))
//         ) : (
//           <p>他のユーザーが見つかりません</p>
//         )}
//       </div> */}

//       {/* 右カラム：チャット画面 */}
//           {/* メインエリア：チャット画面 */}
//           <main className={styles.main}>
//           <div style={{ marginLeft: "100px" }}>
//             {/* <h1>Let's Chat with Your Friends!</h1> */}
//             <div>
//               <h2>トーク</h2>
//               <div style={{ marginBottom: "600px" }}>
//                 {messages.map((msg) => (
//                   <p key={msg.id}>
//                     <strong>{msg.sender}:</strong> {msg.content}
//                   </p>
//                 ))}
//               </div>
//               <input
//                 type="text"
//                 placeholder="メッセージを入力"
//                 value={message}
//                 onChange={(e) => setMessage(e.target.value)}
//                 style={{ marginRight: "100px", padding: "5px" }}
//               />
//               <button onClick={handleSendMessage} className={styles.primary}>
//                 送信
//               </button>
//             </div>
//           </div>
//           </main>
//       <footer className={styles.footer}>
//       <button onClick={handleLogout}>ログアウト</button>
//       </footer>
//     </div>
//   );
// }
