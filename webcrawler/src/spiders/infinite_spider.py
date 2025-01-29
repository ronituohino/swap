from scrapy.spiders import CrawlSpider, Rule
from scrapy.linkextractors import LinkExtractor
from ..items import WebcrawlerItem


class InfiniteSpider(CrawlSpider):
	name = "infinite"

	def __init__(
		self, start_urls="http://www.example.com", categories="", *args, **kwargs
	):
		self.start_urls = start_urls.split("|")
		self.categories = {}
		if categories:
			for category in categories.split("|"):
				if ":" in category:
					cat_name, keywords = category.split(":", 1)
					self.categories[cat_name] = keywords.split(",")
				else:
					self.categories[category] = []

		self.logger.info("Running with args:")
		self.logger.info(f"- start_urls: {self.start_urls}")
		self.logger.info(f"- categories: {self.categories}")

		super().__init__(*args, **kwargs)

	rules = (Rule(LinkExtractor(allow=()), callback="parse_item", follow=True),)

	def text_contains_keywords(self, text, keywords):
		if not keywords:
			return True
		return any(keyword.lower() in text.lower() for keyword in keywords)

	def get_visible_text(self, response):
		selector = "p::text, h1::text, h2::text, h3::text, h4::text, div::text"

		text = " ".join(t.strip() for t in response.css(selector).getall() if t.strip())
		return text

	def parse_item(self, response):
		text = self.get_visible_text(response)
		for category, keywords in self.categories.items():
			if self.text_contains_keywords(text, keywords):
				item = WebcrawlerItem(url=response.url, category=category)

				self.logger.info(f"Parse item: {item}")

				yield item
