NODE_0_LOG="./node_0.log"
NODE_1_LOG="./node_1.log"
NODE_2_LOG="./node_2.log"

rm -rf ./node_*.log

function quoted {
  SPACE=""
  for arg in "$@"; do
    printf "%s" "$SPACE\"$arg\""
    SPACE=" "
  done
  printf "\n"
}

unset TMUX
tmux new-session -d -s "demo" \; \
  new-window   -t "demo" \; \
  \
  split-window -t "demo:0" -v \; \
  split-window -t "demo:0.0" -h \; \
  \
  send-keys -t "demo:0.0" "go run . $(quoted "$@") 0 2>&1 | tee $NODE_0_LOG" Enter \; \
  send-keys -t "demo:0.1" "go run . $(quoted "$@") 1 2>&1 | tee $NODE_1_LOG" Enter \; \
  send-keys -t "demo:0.2" "go run . $(quoted "$@") 2 2>&1 | tee $NODE_2_LOG" Enter \; \
  attach-session -t "demo:0.0"
