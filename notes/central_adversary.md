```mermaid
flowchart LR;

cc00{{Cortex Creeper Node 0}}
cc01{{Cortex Creeper Node 1}}
cc02{{Cortex Creeper Node 2}}
cc03{{Cortex Creeper Node 3}}

cc10{{Cortex Creeper Node 0}}
cc11{{Cortex Creeper Node 1}}
cc12{{Cortex Creeper Node 2}}
cc13{{Cortex Creeper Node 3}}

subgraph ca[Central Adversary]

subgraph bb[Orchestrator]
direction TB;

subgraph bbrt0[Plugin Runtime]
bb0[Byzantine Behavior Plugin A]
end

subgraph bbrt1[Plugin Runtime]
bb1[Byzantine Behavior Plugin B]
end

subgraph bbrt2[Plugin Runtime]
bb2[Byzantine Behavior Plugin C]
end

bbrt0 --> bbrt1
bbrt1 --> bbrt2

end

end

cc00 --(events)--> bb
cc01 --(events)--> bb
cc02 --(events)--> bb
cc03 --(events)--> bb

bb --(events)--> cc10
bb --(events)--> cc11
bb --(events)--> cc12
bb --(events)--> cc13

```
