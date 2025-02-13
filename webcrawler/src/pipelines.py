# Define your item pipelines here
#
# Don't forget to add your pipeline to the ITEM_PIPELINES setting
# See: https://docs.scrapy.org/en/latest/topics/item-pipeline.html


# useful for handling different item types with a single interface
from itemadapter import ItemAdapter
from scrapy.exceptions import DropItem
import pika
import os
import json
import math


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

			# float, ranges from 1 to +inf
			relevance = item["relevances"][word]
			if relevance < 4.0:
				# if the word doesn't have much relevance in the page, just exclude it
				continue
			# adjust relevance to between 1 and 1.8 (from 4.0 to 160)
			relevance = math.log(relevance, 100) + 0.7
			if relevance > 1.8:
				relevance = 1.8

			word_properties["relevance"] = relevance

			# float, ranges from 0 to 1
			term_frequency = item["counts"][word] / item["total_words"]
			word_properties["term_frequency"] = term_frequency

			message["keywords"][word] = word_properties

		self.channel.basic_publish(
			exchange="", routing_key="scraped_items", body=json.dumps(message)
		)
		return item
