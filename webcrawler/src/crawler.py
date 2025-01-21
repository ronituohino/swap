import requests
from bs4 import BeautifulSoup
import argparse
import threading
from queue import Queue
from threading import Lock
import time


def normalize_url(url):
	return url.rstrip("/")


class Worker(threading.Thread):
	def __init__(self, queue, visited, visited_lock, max_depth):
		threading.Thread.__init__(self)
		self.queue = queue
		self.visited = visited
		self.visited_lock = visited_lock
		self.max_depth = max_depth

	def run(self):
		while True:
			item = self.queue.get()
			if item is None:
				break

			url, depth = item
			self.process_url(url, depth)
			self.queue.task_done()

	def process_url(self, current_url, depth):
		if depth > self.max_depth:
			return

		normalized_url = normalize_url(current_url)

		with self.visited_lock:
			if normalized_url in self.visited:
				return
			self.visited.add(normalized_url)

		try:
			response = requests.get(current_url)
			soup = BeautifulSoup(response.text, "html.parser")
			print(f"Crawling: {current_url} (depth: {depth})")

			if depth < self.max_depth:
				links = soup.find_all("a")
				for link in links:
					href = link.get("href")
					if href and href.startswith("http"):
						self.queue.put((href, depth + 1))
		except requests.RequestException:
			pass


def crawl(url, num_threads=5, max_depth=2):
	queue = Queue()
	visited = set()
	visited_lock = Lock()

	workers = []
	for _ in range(num_threads):
		worker = Worker(queue, visited, visited_lock, max_depth)
		worker.start()
		workers.append(worker)

	queue.put((url, 0))

	queue.join()

	for _ in workers:
		queue.put(None)
	for worker in workers:
		worker.join()

	return visited


def main():
	parser = argparse.ArgumentParser(description="Web crawler")
	parser.add_argument("url", help="Starting URL to crawl")
	parser.add_argument("--threads", type=int, default=5, help="Number of threads")
	parser.add_argument("--depth", type=int, default=2, help="Maximum crawl depth")
	args = parser.parse_args()

	start_time = time.time()
	visited_urls = crawl(args.url, args.threads, args.depth)
	end_time = time.time()

	print(f"\nTotal URLs visited: {len(visited_urls)}")
	print(f"Time taken: {end_time - start_time:.2f} seconds")


if __name__ == "__main__":
	main()
