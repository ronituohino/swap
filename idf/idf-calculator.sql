DO $$
DECLARE
    total_websites float;
BEGIN
    RAISE NOTICE 'Staring IDF calculation:';

    -- Get total number of websites
    SELECT COUNT(*)::float INTO total_websites FROM websites;

    -- Calculate IDF values and update relations table directly
    UPDATE relations r
    SET idf = CASE
        WHEN doc_freq > 0 THEN GREATEST(0.01, LN(total_websites / doc_freq::float))
        ELSE 0.01
    END
    FROM (
        SELECT
            r.keyword_id,
            COUNT(DISTINCT r.website_id) AS doc_freq
        FROM relations r
        GROUP BY r.keyword_id
    ) AS keyword_doc_freqs
    WHERE r.keyword_id = keyword_doc_freqs.keyword_id;

    -- Log final statistics
    RAISE NOTICE 'IDF calculation complete. Statistics:';
    RAISE NOTICE 'Min IDF: %', (SELECT MIN(idf) FROM relations WHERE idf > 0);
    RAISE NOTICE 'Max IDF: %', (SELECT MAX(idf) FROM relations);
    RAISE NOTICE 'Avg IDF: %', (SELECT AVG(idf) FROM relations WHERE idf > 0);
    RAISE NOTICE 'Total relations updated: %', (SELECT COUNT(*) FROM relations WHERE idf > 0);

    -- 0.01 IDF values
    RAISE NOTICE 'Number of 0.01 IDF values: %', (SELECT COUNT(*) FROM relations WHERE idf = 0.01);

    -- Print IDF distribution
    RAISE NOTICE 'IDF Distribution:';
    RAISE NOTICE '%', (
      SELECT string_agg(idf_range || ': ' || count || ' (' ||
             ROUND(CAST((count::float / total * 100) AS numeric), 2) || '%)', E'\n')
      FROM (
        SELECT
          CASE
            WHEN idf < 1 THEN 'Very low (0-1)'
            WHEN idf < 2 THEN 'Low (1-2)'
            WHEN idf < 3 THEN 'Medium (2-3)'
            WHEN idf < 4 THEN 'High (3-4)'
            WHEN idf < 5 THEN 'Very High (4-5)'
            ELSE 'Extremely High (5+)'
          END as idf_range,
          COUNT(*) as count,
          SUM(COUNT(*)) OVER () as total
        FROM relations
        GROUP BY
          CASE
            WHEN idf < 1 THEN 'Very low (0-1)'
            WHEN idf < 2 THEN 'Low (1-2)'
            WHEN idf < 3 THEN 'Medium (2-3)'
            WHEN idf < 4 THEN 'High (3-4)'
            WHEN idf < 5 THEN 'Very High (4-5)'
            ELSE 'Extremely High (5+)'
          END
        ORDER BY idf_range
      ) t
    );

END $$;