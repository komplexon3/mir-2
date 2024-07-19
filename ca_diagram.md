flowchart LR;

subgraph n0[Node]
cc0{{Cortex Creeper}}
end
subgraph n1[Node]
cc1{{Cortex Creeper}}
ij1{{Injection Entry}}
ic1{{Interceptor Out}}
end

subgraph ca[Central Adversary]

finj{{Filter injected / non interestin events}} -- (events to 'skip') --> cc1
finj -- (interesting events) --> cn

cn{{Check network events}} -- network events --> as
cn -- other events --> cb{{Check node may \n take byzantine action}}

cb -- byz ok --> as
cb -- byz not ok --> cc1

subgraph as[Action Selection]
direction LR

jf{{Just Forward}}
dd{{Duplicate}}
d{{Drop}}
end

as --> ba{{If action was byzantine, mark node as byzantine}}
ba --> cc1

end

cc0 --(events)--> finj

cc1 -- inject --> ij1
cc1 -- forward --> ic1
