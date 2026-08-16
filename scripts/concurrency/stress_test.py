import socket
import urllib.request
import urllib.error
import concurrent.futures
import time

UDP_IP = "127.0.0.1"
UDP_PORT = 8125
HTTP_URL = "http://127.0.0.1:8080/ingest"

def blast_udp(batch_size):
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    for i in range(batch_size):
        msg = f"memory.leak.count:{i}|c".encode('utf-8')
        sock.sendto(msg, (UDP_IP, UDP_PORT))

def blast_http(worker_id, batch_size):
    http_429_count = 0
    http_202_count = 0

    for i in range(batch_size):
        try:
            payload = f'{{"alert": "OOM_Warning", "worker": {worker_id}, "id": {i}}}'.encode('utf-8')
            req = urllib.request.Request(HTTP_URL, data=payload, method='POST')

            with urllib.request.urlopen(req, timeout=1) as response:
                if response.getcode() == 202:
                    http_202_count += 1

        except urllib.error.HTTPError as e:
            if e.code == 429:
                http_429_count += 1 # Go edge rejected us
        except Exception:
            pass # ignore standard TCP resets during tests

    return http_202_count, http_429_count

if __name__ == "__main__":
    print("Starting Multi-Protocol Stress Test...")
    start_time = time.time()

    with concurrent.futures.ThreadPoolExecutor(max_workers=10) as executor:

        # 15.000 UDP packets
        executor.submit(blast_udp, 15000)

        # blasting 500 HTTP requests
        http_futures = [executor.submit(blast_http, i, 500) for i in range(5)]

        total_202 = 0
        total_429 = 0

        for future in concurrent.futures.as_completed(http_futures):
            success, rejected = future.result()
            total_202 += success
            total_429 += rejected

    duration = time.time() - start_time
    print(f"Test completed in {duration:.2f} seconds.")
    print(f"HTTP Results -> Accepted: {total_202}, Shed/Rejected (HTTP 429): {total_429}")
    print("Check Go Edge terminal for UDP drop logs.")