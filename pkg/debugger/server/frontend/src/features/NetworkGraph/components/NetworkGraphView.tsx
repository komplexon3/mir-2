import { NetworkGraph } from "./NetworkGraph";
import { MirEdgeType, MirNodeType } from "../types";

const nodes: MirNodeType[] = [
  { label: "1", id: "1" },
  { label: "2", id: "2" },
  { label: "3", id: "3" },
  { label: "4", id: "4" },
];

const edges: MirEdgeType[] = [
  { sourceId: "1", targetId: "2" },
  { sourceId: "1", targetId: "3" },
  { sourceId: "1", targetId: "4" },
  { sourceId: "2", targetId: "1" },
  { sourceId: "2", targetId: "3" },
  { sourceId: "2", targetId: "4" },
  { sourceId: "3", targetId: "1" },
  { sourceId: "3", targetId: "2" },
  { sourceId: "3", targetId: "4" },
];

export const NetworkGraphView = () => (
  <div className="h-screen w-screen">
    <NetworkGraph nodes={nodes} edges={edges} />
  </div>
);
