import "@/styles/globals.css";
import type { AppProps } from "next/app";
//import { WebSocketProvider } from "@/pages/WebSocketContext";

export default function App({ Component, pageProps }: AppProps) {

  return <Component {...pageProps} />;

  // return (
  //   <WebSocketProvider>
  //     <Component {...pageProps} />
  //   </WebSocketProvider>
  // );

}
