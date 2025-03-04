from scrapy.spiders import CrawlSpider, Rule
from scrapy.linkextractors import LinkExtractor
from ..items import WebcrawlerItem
from random import shuffle

import json


def read_resource(filename: str, top_type: type):
	with open(f"resources/{filename}.json", "r") as file:
		data = json.load(file)
		if top_type is dict and isinstance(data, dict):
			# dict
			return data
		elif hasattr(data, "__len__") and (not isinstance(data, str)):
			# list
			return data
		else:
			raise ValueError(f"{filename}.json not properly loaded")


class InfiniteSpider(CrawlSpider):
	name = "infinite"

	def __init__(self, *args, **kwargs):
		start_urls = read_resource("start_sites", list)
		shuffle(start_urls)
		self.start_urls = start_urls

		lang = read_resource("languages", list)
		self.languages = set(lang)

		self.selectors = read_resource("selectors", dict)

		chars = read_resource("chars_to_include", list)
		self.chars_to_include = set(chars)

		words = read_resource("words_to_delete", list)
		self.words_to_delete = set(words)

		self.lemmatize = read_resource("lemmatize", dict)
		self.transforms = read_resource("transforms", dict)

		super().__init__(*args, **kwargs)

	rules = (Rule(LinkExtractor(allow=()), callback="parse_item", follow=True),)

	def parse_item(self, response):
		keywords = set()
		counts = {}
		relevances = {}
		# some words could be counted multiple times if in mulitple selectors
		# e.g a word in <body><main><h1> is counted 3 times
		total_words = 0

		title = response.css("title::text").get()
		language = response.css("html::attr(lang)").get()

		# some sites do not specify language (bad you!), but index them anyways
		if language in self.languages or language is None:
			for selector, semantic_relevance in self.selectors.items():
				texts = response.css(selector).getall()
				for text in texts:
					# strip() removes excess whitespace on both sides
					# lower() makes all characters lowercase
					# split() splits text by space and newline
					words = text.strip().lower().split()
					word_amount = len(words)

					for index, word in enumerate(words):
						# remove english possessive from end
						if word.endswith("'s"):
							word = word.rstrip("'s")

						# remove special characters from word
						for ch in word:
							if ch not in self.chars_to_include:
								word = word.replace(ch, "")

						# don't index words that are only 1 character long
						if len(word) < 2:
							continue

						# lemmatize words to improve search (e.g. "cars" => "car")
						if word in self.lemmatize:
							word = self.lemmatize[word]

						# remove very generic words that don't provide insight
						if word in self.words_to_delete:
							continue

						# apply word transforms to improve search and optimize space (e.g. "one" => "1")
						if word in self.transforms:
							word = self.transforms[word]

						# words that are early in text are more relevant
						# first 10 words in any text have boosted positional_relevance
						# then the positional_relevance approaches 1, speed depending on text length
						offset_index = index - 10
						if offset_index < 0:
							offset_index = 0
						# ranges from 2 to 1
						positional_relevance = (1 - (offset_index / word_amount)) + 1

						total_relevance = semantic_relevance * positional_relevance

						if word in keywords:
							counts[word] += 1
							relevances[word] += total_relevance
						else:
							keywords.add(word)
							counts[word] = 1
							relevances[word] = total_relevance

						total_words += 1

		item = WebcrawlerItem(
			url=response.url,
			title=title,
			language=language,
			keywords=list(keywords),
			counts=counts,
			relevances=relevances,
			total_words=total_words,
		)
		yield item
