import ReactFlow, { Node } from "reactflow";
import { MirNode } from "./MirNode";
import { MirNodeType, MirEdgeType } from "../types";
import Dagre from "@dagrejs/dagre";

import "reactflow/dist/style.css";

const g = new Dagre.graphlib.Graph().setDefaultEdgeLabel(() => ({}));

const nodeTypes = { mirNode: MirNode };

function getLayoutedElements(
  nodes: any,
  edges: any,
  options: { direction: string }
) {
  g.setGraph({ rankdir: options.direction });

  edges.forEach((edge: any) => g.setEdge(edge.source, edge.target));
  nodes.forEach((node: any) => g.setNode(node.id, node));

  Dagre.layout(g);

  return {
    nodes: nodes.map((node: any) => {
      const { x, y } = g.node(node.id);

      return { ...node, position: { x, y } };
    }),
    edges,
  };
}

interface NetworkGraphProps {
  nodes: MirNodeType[];
  edges: MirEdgeType[];
}

export function NetworkGraph(props: NetworkGraphProps) {
  const { nodes, edges } = props;
  const flowNodes: Node<MirNodeType>[] = nodes.map((node) => ({
    id: node.id,
    data: node,
    type: "mirNode",
    position: { x: 0, y: 0 },
    width: 100,
    height: 100,
  }));

  const flowEdges = edges.map((edge) => ({
    id: `${edge.sourceId}-${edge.targetId}`,
    source: edge.sourceId,
    target: edge.targetId,
  }));

  const layoutedElements = getLayoutedElements(flowNodes, flowEdges, {
    direction: "TB",
  });

  return (
    <div className="h-full w-full">
      <ReactFlow {...layoutedElements} nodeTypes={nodeTypes} />;
    </div>
  );
}
