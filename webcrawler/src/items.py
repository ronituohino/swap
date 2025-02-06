# Define here the models for your scraped items
#
# See documentation in:
# https://docs.scrapy.org/en/latest/topics/items.html

from scrapy.item import Item, Field


class WebcrawlerItem(Item):
	url = Field()
	keywords = Field()
	counts = Field()
	relevances = Field()
	terms = Field()
