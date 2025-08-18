#!/usr/bin/env python3
"""
Script to load users from PostgreSQL to Valkey for authentication cache.
Uses base64(email:passwordHash) as key and userID as value in users:auth hash.
"""

import os
import sys
import base64
import psycopg2
import redis
from psycopg2.extras import RealDictCursor


def get_db_connection():
    """Get PostgreSQL database connection."""
    try:
        conn = psycopg2.connect(
            host=os.environ.get('DB_HOST', 'localhost'),
            port=os.environ.get('DB_PORT', '5432'),
            database=os.environ.get('DB_NAME', 'bulbul'),
            user=os.environ.get('DB_USER', 'postgres'),
            password=os.environ.get('DB_PASSWORD', ''),
            cursor_factory=RealDictCursor
        )
        return conn
    except psycopg2.Error as e:
        print(f"Error connecting to PostgreSQL: {e}")
        sys.exit(1)


def get_valkey_connection():
    """Get Valkey/Redis connection."""
    try:
        valkey_host = os.environ.get('VALKEY_HOST', 'localhost')
        valkey_port = int(os.environ.get('VALKEY_PORT', '6379'))
        valkey_password = os.environ.get('VALKEY_PASSWORD', '')
        
        r = redis.Redis(
            host=valkey_host,
            port=valkey_port,
            password=valkey_password if valkey_password else None,
            decode_responses=True
        )
        # Test connection
        r.ping()
        return r
    except redis.RedisError as e:
        print(f"Error connecting to Valkey: {e}")
        sys.exit(1)


def load_users_to_valkey():
    """Load users from PostgreSQL to Valkey hash."""
    db_conn = get_db_connection()
    valkey_conn = get_valkey_connection()
    
    hash_key = os.environ.get('VALKEY_USERS_HASH_KEY', 'users:auth')
    page_size = 10000
    offset = 0
    total_loaded = 0
    
    print(f"Starting user load to Valkey hash: {hash_key}")
    
    try:
        cursor = db_conn.cursor()
        
        # Clear existing hash
        valkey_conn.delete(hash_key)
        print(f"Cleared existing hash: {hash_key}")
        
        while True:
            # Fetch users in batches
            query = """
                SELECT user_id, email, password_hash 
                FROM users 
                WHERE is_active = true 
                AND password_hash IS NOT NULL 
                AND password_hash != ''
                ORDER BY user_id 
                LIMIT %s OFFSET %s
            """
            
            cursor.execute(query, (page_size, offset))
            users = cursor.fetchall()
            
            if not users:
                break
            
            # Prepare batch data for hash
            hash_data = {}
            for user in users:
                # Create cache key: base64(email:passwordHash)
                auth_string = f"{user['email']}:{user['password_hash']}"
                cache_key = base64.b64encode(auth_string.encode('utf-8')).decode('utf-8')
                hash_data[cache_key] = str(user['user_id'])
            
            # Batch insert to hash
            if hash_data:
                valkey_conn.hset(hash_key, mapping=hash_data)
                total_loaded += len(hash_data)
                print(f"Loaded batch of {len(hash_data)} users (total: {total_loaded})")
            
            offset += page_size
        
        print(f"Successfully loaded {total_loaded} users to Valkey")
        
    except Exception as e:
        print(f"Error loading users: {e}")
        sys.exit(1)
    finally:
        cursor.close()
        db_conn.close()
        valkey_conn.close()


if __name__ == "__main__":
    load_users_to_valkey()