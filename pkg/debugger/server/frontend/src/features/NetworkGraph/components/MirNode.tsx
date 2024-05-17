import { Handle, Position, NodeProps } from "reactflow";
import { MirNodeType } from "../types";

export function MirNode(props: NodeProps<MirNodeType>) {
  const { data } = props;
  return (
    <div className="rounded-full bg-slate-400 border-2 border-slate-600 w-8 h-8 flex items-center justify-center">
      <div className="text-white">{data.label}</div>
      <Handle
        type="source"
        position={Position.Left}
        className="absolute h-full w-full opacity-0"
      />
      <Handle
        type="target"
        position={Position.Left}
        className="absolute h-full w-full opacity-0"
      />
    </div>
  );
}
