import {
  For,
  Match,
  Suspense,
  Switch,
  createEffect,
  createResource,
  createSignal,
  onMount,
} from "solid-js";

import { Spinner } from "./Spinner.tsx";

import "./Search.css";
import { Icon } from "./Icon.tsx";

const SEARCH_API_URL = "https://api-nqknowvima-lz.a.run.app/search";

type SearchResults = {
  results: SearchHits[];
  queryTime: number;
  totalHits: number;
};

type SearchHits = {
  title: string;
  url: string;
  score: number;
  keywords: string[];
};

const getInitialQuery = () => {
  const searcParams = new URLSearchParams(window.location.search);
  const query = searcParams.get("q");
  return query || undefined;
};

// Search result fetching from API
const fetchSearchResults = async (
  search: string | undefined
): Promise<SearchResults | undefined> => {
  const trimmed = search?.trim();
  if (!trimmed) {
    // Empty or undefined
    return;
  }

  // Update query param
  const url = new URL(window.location.href);
  url.searchParams.set("q", trimmed);
  window.history.pushState(null, "", url.toString());

  // Fetch results from API
  const result = await fetch(
    `${SEARCH_API_URL}?q=${encodeURIComponent(trimmed)}`
  );
  const data = await result.json();

  if (
    !data ||
    !data.query_time ||
    !data.results ||
    !data.total_hits ||
    !Array.isArray(data.results)
  ) {
    return {
      queryTime: 0,
      results: [],
      totalHits: 0,
    };
  }

  return {
    queryTime: data.query_time,
    results: data.results.map((result: any) => ({
      url: result.url,
      title: result.title,
      score: result.score,
      keywords: result.keywords,
    })),
    totalHits: data.total_hits,
  } satisfies SearchResults;
};

const isEmptyResult = (searchResults: SearchResults | undefined) => {
  if (searchResults === undefined) {
    return false;
  }
  if (searchResults.totalHits === 0) {
    return true;
  }
};

export function Search() {
  let formRef: HTMLFormElement | undefined;
  const [search, setSearch] = createSignal<undefined | string>(
    getInitialQuery()
  );
  const [searchResults] = createResource(search, fetchSearchResults);

  // Set q search param to input if provided on page load
  onMount(() => {
    const q = getInitialQuery();
    if (q && formRef) {
      const queryInput = formRef.querySelector(
        "input[name=query]"
      ) as HTMLInputElement | null;
      if (queryInput) {
        queryInput.value = q;
      }
    }
  });

  return (
    <main>
      <form
        ref={formRef}
        onSubmit={(e) => {
          e.preventDefault();
          const data = new FormData(e.target as HTMLFormElement);
          const query = data.get("query")?.toString();
          if (query) {
            setSearch(query);
          }
        }}
        class="search-bar"
      >
        <input type="text" name="query" class="search-input" />
        <button class="button" type="submit" aria-label="Search">
          <Icon />
        </button>
      </form>

      <div class="search-results">
        <Switch>
          <Match when={searchResults.error}>
            <span>Error: {searchResults.error.message}</span>
          </Match>
          <Match when={searchResults.loading}>
            <Spinner class="spinner" />
          </Match>
          <Match when={isEmptyResult(searchResults())}>
            <span class="no-result">No results found.</span>
          </Match>
          <Match
            when={
              searchResults() !== undefined && !isEmptyResult(searchResults())
            }
          >
            <span class="query-details">
              query time: {searchResults()?.queryTime} | total hits:{" "}
              {searchResults()?.totalHits}
            </span>
            <ul class="result-list">
              <For each={searchResults()?.results}>
                {(hit) => (
                  <li>
                    <a href={hit.url} target="_blank" class="result">
                      <span class="result-title">{hit.title}</span>
                      <span class="result-details">
                        ({hit.url}) <br /> {hit.keywords.join(", ")} |{" "}
                        {hit.score}
                      </span>
                    </a>
                  </li>
                )}
              </For>
            </ul>
          </Match>
        </Switch>
      </div>
    </main>
  );
}
