# TCP Server

## Description
TCP Server is a small concurrently safe server that responds to requests from other services. The server listens to requests from localhost:3000 on a raw tcp socket and executes commands received from a remote client. 

## Installation and Setup
1. Ensure you have Go (v1.22.0) installed on your system. You can install the latest version [here](https://go.dev/doc/install) 
2. Clone this repository:
   ```
   https://github.com/Oyal2/tcp-server.git
   ```
3. Navigate to the project directory:
   ```
   cd tcp-server
   ```
4. Make sure the unit tests pass:
   ```
    make test
   ```
5. Build the project:
   ```
   make build
   ```

## Usage
1. To start the server you can run:
   ```
   ./tcp-server
   ```

    or 

    ```
    make run
    ```
2. The server will start listening on localhost:3000
3. Send task requests to the server using a TCP client in the following JSON format:
   ```json
   {
     "command": ["./cmd", "--flag", "argument1", "argument2"],
     "timeout": 500
   }
   ```
4. The server will respond with a JSON result containing the task's result

## Technical Details

### Task Request Structure
Each incoming task request follows this JSON structure:
```json
{
  "command": ["./cmd", "--flag", "argument1", "argument2"],
  "timeout": 500
}
```
- `command`: ARGV array of arguments, the first of which is the absolute
path to the command being executed.
- `timeout`: Time in milliseconds after a task should be terminated. 
  - A timeout of 0 or a missing timeout field means there is no timeout.

### Task Result Structure
The server responds with a task result in the following JSON format:
```json
{
  "command": ["./cmd", "--flag", "argument1", "argument2"],
  "executed_at": 1621234567,
  "duration_ms": 123,
  "exit_code": 0,
  "output": "Command output here",
  "error": ""
}
```
- `executed_at`: Unix timestamp when the task was executed.
- `duration_ms`: Execution time in milliseconds.
- `exit_code`: The exit status of the subprocess.
  - `-1` if the process failed to execute or timeout was exceeded.
- `output`: Everything written to STDOUT.
- `error`: Error message if any

### Timeout Handling
- If the specified timeout is exceeded, the task is terminated.
- In case of a timeout, the `exit_code` is set to -1 and the `error` field contains "timeout exceeded".