#!/usr/bin/env python3
"""
Elasticsearch Events Data Loader
Loads events from JSON file into Elasticsearch with Russian language support
"""

import json
import os
import sys
import time
from datetime import datetime
from elasticsearch import Elasticsearch, helpers


def main():
    # Configuration from environment
    elasticsearch_url = os.getenv('ELASTICSEARCH_URL', 'http://localhost:9200')
    index_name = os.getenv('ELASTICSEARCH_INDEX', 'events')
    data_file = os.getenv('DATA_FILE', '/app/data/events.json')
    batch_size = int(os.getenv('BATCH_SIZE', '1000'))

    print(f"Starting Elasticsearch events loader...")
    print(f"Elasticsearch URL: {elasticsearch_url}")
    print(f"Index: {index_name}")
    print(f"Data file: {data_file}")
    print(f"Batch size: {batch_size}")

    # Initialize Elasticsearch client
    try:
        es = Elasticsearch(
            [elasticsearch_url],
            verify_certs=False,
            ssl_show_warn=False,
            request_timeout=30,
            max_retries=3,
            retry_on_timeout=True
        )
        
        # Check connection
        if not es.ping():
            print("ERROR: Cannot connect to Elasticsearch")
            sys.exit(1)
            
        print("✓ Connected to Elasticsearch")
    except Exception as e:
        print(f"ERROR: Failed to connect to Elasticsearch: {e}")
        sys.exit(1)

    # Wait for Elasticsearch to be fully ready
    print("Waiting for Elasticsearch cluster to be ready...")
    for attempt in range(30):
        try:
            health = es.cluster.health(wait_for_status='yellow', timeout='10s')
            if health['status'] in ['yellow', 'green']:
                print(f"✓ Elasticsearch cluster is {health['status']}")
                break
        except Exception as e:
            print(f"Attempt {attempt + 1}/30: Waiting for cluster health... ({e})")
            time.sleep(2)
    else:
        print("ERROR: Elasticsearch cluster not ready")
        sys.exit(1)

    # Create index with Russian analyzer if it doesn't exist
    index_settings = {
        "settings": {
            "number_of_shards": 1,
            "number_of_replicas": 0,
            "analysis": {
                "analyzer": {
                    "russian_analyzer": {
                        "type": "custom",
                        "tokenizer": "standard",
                        "filter": ["lowercase", "russian_stop", "russian_stemmer"]
                    }
                },
                "filter": {
                    "russian_stop": {
                        "type": "stop",
                        "stopwords": "_russian_"
                    },
                    "russian_stemmer": {
                        "type": "stemmer",
                        "language": "russian"
                    }
                }
            }
        },
        "mappings": {
            "properties": {
                "id": {"type": "long"},
                "title": {
                    "type": "text",
                    "analyzer": "russian_analyzer",
                    "fields": {
                        "keyword": {
                            "type": "keyword",
                            "ignore_above": 256
                        }
                    }
                },
                "description": {
                    "type": "text",
                    "analyzer": "russian_analyzer"
                },
                "type": {"type": "keyword"},
                "datetime_start": {
                    "type": "date",
                    "format": "strict_date_optional_time||epoch_millis"
                },
                "provider": {"type": "keyword"},
                "external": {"type": "boolean"},
                "total_seats": {"type": "integer"},
                "created_at": {"type": "date"},
                "updated_at": {"type": "date"}
            }
        }
    }

    try:
        # Delete existing index if it exists (to ensure clean data with correct format)
        if es.indices.exists(index=index_name):
            print(f"Deleting existing index: {index_name}")
            es.indices.delete(index=index_name)
            print("✓ Existing index deleted")
            
        print(f"Creating index: {index_name}")
        es.indices.create(index=index_name, body=index_settings)
        print("✓ Index created successfully")
    except Exception as e:
        print(f"ERROR: Failed to manage index: {e}")
        sys.exit(1)

    # Load and process events data
    try:
        print(f"Loading events from: {data_file}")
        
        if not os.path.exists(data_file):
            print(f"ERROR: Data file not found: {data_file}")
            sys.exit(1)
            
        with open(data_file, 'r', encoding='utf-8') as f:
            events_data = json.load(f)
            
        print(f"✓ Loaded {len(events_data)} events from JSON file")
        
    except Exception as e:
        print(f"ERROR: Failed to load events data: {e}")
        sys.exit(1)

    # Prepare documents for bulk indexing
    def prepare_documents():
        for event in events_data:
            # Convert datetime strings to proper format
            if 'datetime_start' in event and isinstance(event['datetime_start'], str):
                try:
                    # Try parsing different datetime formats
                    dt = None
                    for fmt in ['%Y-%m-%d %H:%M:%S', '%Y-%m-%dT%H:%M:%S', '%Y-%m-%dT%H:%M:%SZ']:
                        try:
                            dt = datetime.strptime(event['datetime_start'], fmt)
                            break
                        except ValueError:
                            continue
                    
                    if dt:
                        event['datetime_start'] = dt.strftime('%Y-%m-%dT%H:%M:%SZ')
                except Exception as e:
                    print(f"Warning: Failed to parse datetime for event {event.get('id', 'unknown')}: {e}")
            
            # Add created_at and updated_at timestamps
            now = datetime.now().strftime('%Y-%m-%dT%H:%M:%SZ')
            if 'created_at' not in event or not event['created_at']:
                event['created_at'] = now
            if 'updated_at' not in event or not event['updated_at']:
                event['updated_at'] = now
            
            # Prepare document for Elasticsearch
            doc = {
                "_index": index_name,
                "_id": str(event['id']),
                "_source": event
            }
            
            yield doc

    # Bulk index documents
    try:
        print(f"Starting bulk indexing with batch size: {batch_size}")
        start_time = time.time()
        
        success_count = 0
        error_count = 0
        
        # Use bulk helper for efficient indexing
        for success, info in helpers.parallel_bulk(
            es,
            prepare_documents(),
            index=index_name,
            chunk_size=batch_size,
            thread_count=2,
            request_timeout=60
        ):
            if success:
                success_count += 1
            else:
                error_count += 1
                print(f"Indexing error: {info}")
            
            # Show progress every 10000 documents
            if (success_count + error_count) % 10000 == 0:
                elapsed = time.time() - start_time
                rate = (success_count + error_count) / elapsed
                print(f"Progress: {success_count + error_count} documents processed "
                      f"({success_count} success, {error_count} errors) - {rate:.1f} docs/sec")

        elapsed = time.time() - start_time
        
        print(f"\n=== Indexing Summary ===")
        print(f"Total events processed: {success_count + error_count}")
        print(f"Successfully indexed: {success_count}")
        print(f"Errors: {error_count}")
        print(f"Total time: {elapsed:.2f} seconds")
        print(f"Average rate: {(success_count + error_count) / elapsed:.1f} documents/second")
        
        if error_count > 0:
            print(f"WARNING: {error_count} events failed to index")
        
    except Exception as e:
        print(f"ERROR: Bulk indexing failed: {e}")
        sys.exit(1)

    # Refresh index to make documents searchable immediately
    try:
        print("Refreshing index...")
        es.indices.refresh(index=index_name)
        print("✓ Index refreshed")
    except Exception as e:
        print(f"Warning: Failed to refresh index: {e}")

    # Verify indexing
    try:
        print("Verifying indexing...")
        doc_count = es.count(index=index_name)['count']
        print(f"✓ Index contains {doc_count} documents")
        
        if doc_count == success_count:
            print("✓ All documents indexed successfully!")
        else:
            print(f"WARNING: Expected {success_count} documents, but found {doc_count}")
            
    except Exception as e:
        print(f"Warning: Failed to verify indexing: {e}")

    # Test search functionality
    try:
        print("Testing search functionality...")
        
        # Test basic search
        test_query = {
            "query": {
                "match_all": {}
            },
            "size": 1
        }
        
        result = es.search(index=index_name, body=test_query)
        if result['hits']['total']['value'] > 0:
            sample_event = result['hits']['hits'][0]['_source']
            print(f"✓ Search test successful. Sample event: {sample_event.get('title', 'N/A')}")
        else:
            print("WARNING: No documents found in search test")
            
    except Exception as e:
        print(f"Warning: Search test failed: {e}")

    print("\n🎉 Events loading completed successfully!")


if __name__ == "__main__":
    main()