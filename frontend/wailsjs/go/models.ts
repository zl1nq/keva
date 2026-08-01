export namespace app {
	
	export class AccountDetail {
	    id: string;
	    title: string;
	    username: string;
	    password: string;
	    url: string;
	    note: string;
	    created_at: number;
	    updated_at: number;
	
	    static createFrom(source: any = {}) {
	        return new AccountDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.url = source["url"];
	        this.note = source["note"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class AccountInput {
	    title: string;
	    username: string;
	    password: string;
	    url: string;
	    note: string;
	
	    static createFrom(source: any = {}) {
	        return new AccountInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.url = source["url"];
	        this.note = source["note"];
	    }
	}
	export class AccountSummary {
	    id: string;
	    title: string;
	    username: string;
	    url: string;
	    created_at: number;
	    updated_at: number;
	
	    static createFrom(source: any = {}) {
	        return new AccountSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.username = source["username"];
	        this.url = source["url"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class InitializeVaultInput {
	    master_password: string;
	
	    static createFrom(source: any = {}) {
	        return new InitializeVaultInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.master_password = source["master_password"];
	    }
	}
	export class LockState {
	    initialized: boolean;
	    locked: boolean;
	    auto_lock_minutes: number;
	
	    static createFrom(source: any = {}) {
	        return new LockState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.initialized = source["initialized"];
	        this.locked = source["locked"];
	        this.auto_lock_minutes = source["auto_lock_minutes"];
	    }
	}
	export class PasswordOptions {
	    length: number;
	    include_uppercase: boolean;
	    include_lowercase: boolean;
	    include_numbers: boolean;
	    include_symbols: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PasswordOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.length = source["length"];
	        this.include_uppercase = source["include_uppercase"];
	        this.include_lowercase = source["include_lowercase"];
	        this.include_numbers = source["include_numbers"];
	        this.include_symbols = source["include_symbols"];
	    }
	}
	export class RuntimeStatus {
	    quick_search_shortcut: string;
	    tray_available: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RuntimeStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.quick_search_shortcut = source["quick_search_shortcut"];
	        this.tray_available = source["tray_available"];
	    }
	}
	export class Settings {
	    auto_lock_minutes: number;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.auto_lock_minutes = source["auto_lock_minutes"];
	    }
	}
	export class UnlockInput {
	    master_password: string;
	
	    static createFrom(source: any = {}) {
	        return new UnlockInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.master_password = source["master_password"];
	    }
	}

}

