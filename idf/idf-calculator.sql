CREATE OR REPLACE PROCEDURE update_idf_values()
LANGUAGE plpgsql
AS $$
DECLARE
  total_websites float;
  batch_size integer := 1000000;
  max_id integer;
  current_id integer := 0;
  total_batches integer;
  current_batch integer := 1;
BEGIN
  RAISE NOTICE '[%] Starting IDF calculation:', NOW()::timestamp(0);

  SELECT COUNT(*)::float INTO total_websites FROM websites;
  SELECT MAX(id) INTO max_id FROM relations;
  total_batches := CEIL(max_id::float / batch_size);

  WHILE current_id <= max_id LOOP
    RAISE NOTICE '[%] Processing batch % of % (IDs % to %)',
      NOW()::timestamp(0),
      current_batch,
      total_batches,
      current_id,
      LEAST(current_id + batch_size, max_id);

    UPDATE relations r
    SET idf = CASE
      WHEN doc_freq > 0 THEN GREATEST(0.01, LN(total_websites / doc_freq::float))
      ELSE 0.01
    END
    FROM (
      SELECT
        r2.keyword_id,
        COUNT(DISTINCT r2.website_id) AS doc_freq
      FROM relations r2
      GROUP BY r2.keyword_id
    ) AS keyword_doc_freqs
    WHERE r.keyword_id = keyword_doc_freqs.keyword_id
    AND r.id > current_id
    AND r.id <= current_id + batch_size;

    current_id := current_id + batch_size;
    current_batch := current_batch + 1;
    
    COMMIT;
  END LOOP;

  -- Log final statistics
  RAISE NOTICE '[%] IDF calculation complete. Statistics:', NOW()::timestamp(0);
  RAISE NOTICE '[%] Min IDF: %', NOW()::timestamp(0), (SELECT MIN(idf) FROM relations WHERE idf > 0);
  RAISE NOTICE '[%] Max IDF: %', NOW()::timestamp(0), (SELECT MAX(idf) FROM relations);
  RAISE NOTICE '[%] Avg IDF: %', NOW()::timestamp(0), (SELECT AVG(idf) FROM relations WHERE idf > 0);
  RAISE NOTICE '[%] Total relations updated: %', NOW()::timestamp(0), (SELECT COUNT(*) FROM relations WHERE idf > 0);

  -- 0.01 IDF values
  RAISE NOTICE '[%] Number of 0.01 IDF values: %', NOW()::timestamp(0), (SELECT COUNT(*) FROM relations WHERE idf = 0.01);

  -- Print IDF distribution
  RAISE NOTICE '[%] IDF Distribution:', NOW()::timestamp(0);
  RAISE NOTICE '[%] %', NOW()::timestamp(0), (
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

END;
$$;

-- Execute the procedure
CALL update_idf_values();