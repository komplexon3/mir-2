import { useState, } from "react";
import { WebSocketConsole } from "./WebSocketConsole";

export const EventDebugger = () => {
  const [webSocketConsoles, setWebSocketConsoles] = useState({});
  const [port, setPort] = useState("");

  const connectWebSocket = () => {
    if (port === "" || port in webSocketConsoles) {
      alert("Port already connected or empty");
      return;
    }

    console.log("Starting : ", port);
    const newWebSocketConsoles = { ...webSocketConsoles };
    newWebSocketConsoles[port] = <WebSocketConsole key={port} port={port} />;

    setWebSocketConsoles(newWebSocketConsoles);
    setPort("");
  };

  return (
    <div className=".event-debugger">
      <div>
        <label>
        Enter Port Number:
          <input
            type="text"
            value={port}
            onChange={(e) => setPort(e.target.value)}
          />
        </label>
        <button className="button" onClick={connectWebSocket}>Connect</button>
      </div>

      <div className="message-containers" id="message-containers">
        {Object.values(webSocketConsoles)}
      </div>
    </div>
  );
};
