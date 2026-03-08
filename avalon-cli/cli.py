#!/usr/bin/env python3
import json
import os
import shlex
import sys
import urllib.error
import urllib.parse
import urllib.request


BASE_URL = "http://localhost:8080"
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))


def print_json(data: dict):
    print(json.dumps(data, ensure_ascii=False, indent=2))


def http_get_json(path: str, query: dict) -> dict:
    url = f"{BASE_URL}{path}?{urllib.parse.urlencode(query)}"
    req = urllib.request.Request(url=url, method="GET")
    with urllib.request.urlopen(req) as resp:
        return json.loads(resp.read().decode("utf-8"))


def http_post_json(path: str, payload: dict) -> dict:
    url = f"{BASE_URL}{path}"
    data = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(
        url=url,
        data=data,
        method="POST",
        headers={"Content-Type": "application/json"},
    )
    with urllib.request.urlopen(req) as resp:
        body = resp.read().decode("utf-8")
        if not body:
            return {}
        return json.loads(body)


def create_new_game() -> int:
    resp = http_post_json("/games/new", {})
    game_id = resp.get("gameId")
    if not isinstance(game_id, int):
        raise ValueError("invalid /games/new response: missing gameId")
    return game_id


def read_bool(prompt: str) -> bool:
    while True:
        raw = input(prompt).strip().lower()
        if raw in {"1", "true", "t", "yes", "y", "да", "д"}:
            return True
        if raw in {"0", "false", "f", "no", "n", "нет", "н"}:
            return False
        print("Expected true/false (or yes/no, да/нет).")


def read_int(prompt: str) -> int:
    while True:
        raw = input(prompt).strip()
        try:
            return int(raw)
        except ValueError:
            print("Expected integer.")


def read_squad(prompt: str) -> list[int]:
    while True:
        raw = input(prompt).strip()
        if not raw:
            print("Expected at least one number, example: 1,2,3")
            continue
        try:
            return [int(x.strip()) for x in raw.split(",") if x.strip()]
        except ValueError:
            print("Invalid format. Example: 1,2,3")


def read_param(name: str, description: str):
    if name in {"approve", "success"}:
        return read_bool(f"{name} ({description}): ")
    if name == "target":
        return read_int(f"{name} ({description}): ")
    if name == "squad":
        return read_squad(f"{name} ({description}) [example: 1,2]: ")
    value = input(f"{name} ({description}): ")
    if name == "message":
        value = value.strip()
        if value.startswith("./"):
            path = os.path.normpath(os.path.join(SCRIPT_DIR, value[2:]))
            with open(path, "r", encoding="utf-8") as f:
                return f.read()
    return value


def get_state(game_id: int, player_id: int) -> dict:
    return http_get_json("/games/state", {"gameId": game_id, "playerId": player_id})


def run_status(game_id: int, player_id: int):
    state = get_state(game_id, player_id)
    print_json(state)


def run_action(game_id: int, player_id: int):
    state = get_state(game_id, player_id)
    print_json(state)

    required_action = state.get("requiredAction")
    if not required_action:
        print("No requiredAction now.")
        return

    action_player_id = required_action.get("playerId")
    action_name = required_action.get("name")
    params_def = required_action.get("paramsDef") or {}
    if action_player_id != player_id:
        print(f"Current turn belongs to playerId={action_player_id}.")
        return

    params = {}
    for param_name, description in params_def.items():
        params[param_name] = read_param(param_name, description)

    payload = {
        "action": {
            "playerId": player_id,
            "name": action_name,
            "params": params,
        }
    }
    next_state = http_post_json("/games/action", payload)
    print_json(next_state)


def run_prompt(game_id: int, player_id: int):
    state = get_state(game_id, player_id)
    print_json(state)

    payload = {"gameState": state, "playerId": player_id}
    prompt = http_post_json("/games/system-prompt", payload)
    print_json(prompt)


def parse_command(line: str):
    parts = shlex.split(line)
    if not parts:
        return None, None
    if parts[0] in {"help", "h", "?"}:
        return "help", None
    if parts[0] in {"exit", "quit", "q"}:
        return "exit", None
    if len(parts) != 2:
        return "invalid", None
    cmd = parts[0].lower()
    if cmd not in {"status", "action", "prompt"}:
        return "invalid", None
    try:
        player_id = int(parts[1])
    except ValueError:
        return "invalid", None
    return cmd, player_id


def main():
    print("Avalon CLI")
    while True:
        raw = input("Enter gameId or 'new': ").strip().lower()
        if raw == "new":
            try:
                game_id = create_new_game()
                print(f"New game created: gameId={game_id}")
                break
            except Exception as e:
                print(f"Failed to create game: {e}")
                continue
        try:
            game_id = int(raw)
            break
        except ValueError:
            print("Expected integer gameId or 'new'.")

    print("Commands: status <player_id> | action <player_id> | prompt <player_id> | help | exit")
    while True:
        line = input("> ").strip()
        cmd, player_id = parse_command(line)
        if cmd is None:
            continue
        if cmd == "help":
            print("Commands: status <player_id> | action <player_id> | prompt <player_id> | help | exit")
            continue
        if cmd == "exit":
            return
        if cmd == "invalid":
            print("Invalid command. Use: status <id>, action <id>, prompt <id>, help, exit")
            continue

        try:
            if cmd == "status":
                run_status(game_id, player_id)
            elif cmd == "action":
                run_action(game_id, player_id)
            elif cmd == "prompt":
                run_prompt(game_id, player_id)
        except urllib.error.HTTPError as e:
            print(f"HTTP {e.code}: {e.read().decode('utf-8', errors='ignore')}")
        except urllib.error.URLError as e:
            print(f"Server error: {e}")
        except Exception as e:
            print(f"Error: {e}")


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\nStopped.")
        sys.exit(0)
