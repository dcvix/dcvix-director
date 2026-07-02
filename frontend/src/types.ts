export interface Session {
    uid: string;
    id: string;
    owner: string;
    'last-seen': string;
    'creation-time': string;
    status: string;
    type: string;
    'num-of-connections': number;
    server_id: string;
}

export interface Server {
    hostname: string;
    'last-seen': string;
    sessions: Session[];
    cores: number;
    free_memory: number;
    total_memory: number;
    cpu_usage: number;
    load1: number;
    load5: number;
    load15: number;
    tags: string[];
}

export interface Agent {
    guid: string;
    hostname: string;
    state: string;
    created_at: string;
    registered_at: string | null;
    last_seen_at: string | null;
    tags: string[];
}
