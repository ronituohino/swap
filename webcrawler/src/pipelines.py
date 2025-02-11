# Define your item pipelines here
#
# Don't forget to add your pipeline to the ITEM_PIPELINES setting
# See: https://docs.scrapy.org/en/latest/topics/item-pipeline.html


# useful for handling different item types with a single interface
from itemadapter import ItemAdapter
from scrapy.exceptions import DropItem
import pika
import os


class WebcrawlerPipeline:
	def __init__(self):
		credentials = pika.PlainCredentials(
		  os.getenv("RMQ_USER"), os.getenv("RMQ_PASSWORD")
		)

		parameters = pika.ConnectionParameters(
		  host=os.getenv("RMQ_HOST"),
		  port=int(os.getenv("RMQ_PORT")),
		  credentials=credentials,
		)

		self.connection = pika.BlockingConnection(parameters)
		self.channel = self.connection.channel()
		self.channel.queue_declare(queue="scraped_items", durable=True)

	def close_spider(self, spider):
		self.connection.close()

	def process_item(self, item, spider):
		if len(item["keywords"]) == 0:
			raise DropItem("No keywords found, maybe language is not compatible")

		message = {}
		message["url"] = item["url"]
		message["title"] = item["title"]
		message["keywords"] = {}

		for word in item["keywords"]:
			word_properties = {}

			# float, ranges from 0 to 1
			term_frequency = item["counts"][word] / item["total_words"]
			word_properties["term_frequency"] = term_frequency

			# float, ranges from 1 to +inf
			relevance = item["relevances"][word]
			word_properties["relevance"] = relevance

			message["keywords"][word] = word_properties

		# print condensed crawl results
		sorted_dict = {}
		for key in sorted(item["relevances"], key=item["relevances"].get, reverse=True):
			sorted_dict[key] = item["relevances"][key]
		most_relevant = list(sorted_dict.keys())[0:5]

		print(
			f"{str(item['url'])} - {item['total_words']} terms, most relevant: {most_relevant}"
		)
		self.channel.basic_publish(
			exchange="", routing_key="scraped_items", body=str(item)
		)
		return item
