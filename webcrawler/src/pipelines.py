# Define your item pipelines here
#
# Don't forget to add your pipeline to the ITEM_PIPELINES setting
# See: https://docs.scrapy.org/en/latest/topics/item-pipeline.html


# useful for handling different item types with a single interface
from itemadapter import ItemAdapter
import pika
import os


class WebcrawlerPipeline:
	def __init__(self):
		# credentials = pika.PlainCredentials(
		# os.getenv("RMQ_USER"), os.getenv("RMQ_PASSWORD")
		# )

		# parameters = pika.ConnectionParameters(
		# host=os.getenv("RMQ_HOST"),
		# port=int(os.getenv("RMQ_PORT")),
		# credentials=credentials,
		# )

		# self.connection = pika.BlockingConnection(parameters)
		# self.channel = self.connection.channel()
		# self.channel.queue_declare(queue="scraped_items", durable=True)
		pass

	def close_spider(self, spider):
		# self.connection.close()
		pass

	def process_item(self, item, spider):
		print(f"{str(item['url'])} -- found: {item['terms']} terms")
		# self.channel.basic_publish(
		# exchange="", routing_key="scraped_items", body=str(item)
		# )
		return item
