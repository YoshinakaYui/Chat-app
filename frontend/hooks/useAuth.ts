import { useState, useEffect } from "react";

export const useAuth = () => {
  const [username, setUsername] = useState<string | null>(null);

  useEffect(() => {
    const storedUsername = localStorage.getItem("loggedInUser");
    if (storedUsername) {
      setUsername(storedUsername);
      console.log("ログインユーザー取得成功:", storedUsername);
    } else {
      console.warn("ログインユーザーが見つかりません");
    }
  }, []);

  return { username };
};


// export function useAuth() {
//   const [username, setUsername] = useState<string | null>(null);
//   const [userId, setUserId] = useState<string | null>(null);

//   useEffect(() => {
//     const storedUsername = localStorage.getItem("loggedInUser");
//     const storedUserId = localStorage.getItem("userId");

//     if (storedUsername && storedUserId) {
//       setUsername(storedUsername);
//       setUserId(storedUserId);
//     }
//   }, []);

//   return { username, userId };

//   const [username, setUsername] = useState<string | null>(localStorage.getItem("username"));

//   return { username };
// }

// export const useAuth = () => {

// };

