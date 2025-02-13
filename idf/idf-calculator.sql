DO $$
DECLARE
    total_websites float;
    batch_size constant int := 1000;
    current_batch int := 0;
    total_batches int;
    doc_freq_distribution text;
BEGIN
    -- Get total number of websites
    SELECT COUNT(*)::float INTO total_websites FROM websites;
    
    -- Calculate total number of batches needed
    SELECT CEIL(COUNT(*)::float / batch_size) INTO total_batches FROM keywords;
    
    -- Create temp table for document frequencies
    CREATE TEMP TABLE doc_frequencies AS
    SELECT 
        k.id as keyword_id,
        COUNT(DISTINCT r.website_id) as doc_freq
    FROM keywords k
    LEFT JOIN relations r ON k.id = r.keyword_id
    GROUP BY k.id;
    
    -- Create index for better performance
    CREATE INDEX ON doc_frequencies (keyword_id);
    
    -- Process keywords in batches
    FOR current_batch IN 0..total_batches-1 LOOP
        -- Update IDF values for current batch
        UPDATE relations r
        SET idf = CASE 
            WHEN df.doc_freq > 0 THEN GREATEST(0.01, LN(total_websites / df.doc_freq::float))
            ELSE 0.01
        END
        FROM (
            SELECT k.id as kid
            FROM keywords k
            ORDER BY k.id
            OFFSET (current_batch * batch_size)
            LIMIT batch_size
        ) batch_keys
        JOIN doc_frequencies df ON df.keyword_id = batch_keys.kid
        WHERE r.keyword_id = batch_keys.kid;
        
        -- Log progress
        RAISE NOTICE 'Processed batch % of %', current_batch + 1, total_batches;
    END LOOP;
    
    -- Drop temporary table
    DROP TABLE doc_frequencies;
    
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