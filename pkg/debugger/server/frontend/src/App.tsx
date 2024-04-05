import { NetworkGraphView } from "@/features/NetworkGraph";
import { EventDebugger } from "./features/EventDebugger";
import { useState } from "react";
import { cn } from "@/utils/cn";

const Modes = { NETWORK: "NETWORK", EVENT: "EVENT" } as const;
type Modes = (typeof Modes)[keyof typeof Modes];

function App() {
  const [mode, setMode] = useState<Modes>(Modes.NETWORK); // TODO: would be cool to keep in url

  return (
    <div className="h-screen w-screen">
      <div className="w-full flex flex-row bg-gray-300">
        <div
          className={cn("w-full text-center py-2 cursor-pointer", {
            "bg-gray-400": mode === Modes.NETWORK,
          })}
          onClick={() => setMode(Modes.NETWORK)}
        >
          Network God Mode
        </div>
        <div
          className={cn("w-full text-center py-2 cursor-pointer", {
            "bg-gray-400": mode === Modes.EVENT,
          })}
          onClick={() => setMode(Modes.EVENT)}
        >
          Event Debugger
        </div>
      </div>
      <div className="pt-2">
        {mode === Modes.NETWORK ? <NetworkGraphView /> : <EventDebugger />}
      </div>
    </div>
  );
}

export default App;
