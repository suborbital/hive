/**
 * 
 * This file represents the Rust "API" for Reactr Wasm runnables. The functions defined herein are used to exchange data
 * between the host (Reactr, written in Go) and the Runnable (a Wasm module, in this case written in Rust).
 * 
 */

// a small wrapper to hold our dynamic Runnable
struct State <'a> {
    ident: i32,
    runnable: &'a dyn runnable::Runnable
}

// the state that holds the user-provided Runnable and the current ident
static mut STATE: State = State {
    ident: 0,
    runnable: &DefaultRunnable{},
};

pub mod runnable {
    use std::mem;
    use std::slice;
    use super::util;

    extern {
        fn return_result(result_pointer: *const u8, result_size: i32, ident: i32);
        fn return_error(code: i32, result_pointer: *const u8, result_size: i32, ident: i32);
    }

    pub struct RunErr {
        pub code: i32,
        pub message: String,
    }

    impl RunErr {
        pub fn new(code: i32, msg: &str) -> Self {
            RunErr {
                code: code,
                message: String::from(msg)
            }
        }
    }

    pub trait Runnable {
        fn run(&self, input: Vec<u8>) -> Result<Vec<u8>, RunErr>;
    }

    pub fn use_runnable(runnable: &'static dyn Runnable) {
        unsafe {
            super::STATE.runnable = runnable;
        }
    }
    
    #[no_mangle]
    pub extern fn allocate(size: i32) -> *const u8 {
        let mut buffer = Vec::with_capacity(size as usize);

        let buffer_slice = buffer.as_mut_slice();
        let pointer = buffer_slice.as_mut_ptr();

        mem::forget(buffer_slice);
    
        pointer as *const u8
    }
    
    #[no_mangle]
    pub extern fn deallocate(pointer: *const u8, size: i32) {
        unsafe {
            let _ = slice::from_raw_parts(pointer, size as usize);
        }
    }
    
    #[no_mangle]
    pub extern fn run_e(pointer: *const u8, size: i32, ident: i32) {
        unsafe { super::STATE.ident = ident };
    
        // rebuild the memory into something usable
        let in_slice: &[u8] = unsafe { 
            slice::from_raw_parts(pointer, size as usize) 
        };
    
        let in_bytes = Vec::from(in_slice);

        let mut code = 0;
    
        // call the runnable and check its result
        let result: Vec<u8> = unsafe { match super::STATE.runnable.run(in_bytes) {
            Ok(val) => val,
            Err(e) => {
                code = e.code;
                util::to_vec(e.message)
            }
        } };
    
        let result_slice = result.as_slice();
        let result_size = result_slice.len();
    
        // call back to reactr to return the result or error
        unsafe { 
            if code != 0 {
                return_error(code, result_slice.as_ptr() as *const u8, result_size as i32, ident);
            } else {
                return_result(result_slice.as_ptr() as *const u8, result_size as i32, ident);
            }
        }
    }
}

pub mod graqhql {

    extern {
        fn graphql_query(endpoint_pointer: *const u8, endpoint_size: i32, query_pointer: *const u8, query_size: i32, ident: i32) -> i32;
    }

    pub fn query(endpoint: &str, query: &str) -> Result<Vec<u8>,super::runnable::RunErr> {
        // let endpoint_slice = String::from(endpoint);
        // let endpoint_pointer = endpoint_slice.as_ptr();

        // let query_slice = query.as_bytes();
        // let query_pointer = query_slice.as_ptr();

        let endpoint_size = endpoint.len() as i32;
        let query_size = query.len() as i32;

        let result_size = unsafe { graphql_query(endpoint.as_ptr(), endpoint_size, query.as_ptr(), query_size, super::STATE.ident) };

        // retreive the result from the host and return it
        match super::ffi::result(result_size) {
            Ok(res) => Ok(res),
            Err(e) => {
                Err(super::runnable::RunErr::new(e.code, "failed to fetch_url"))
            }
        }
    }
}
pub mod http {
    use std::collections::BTreeMap;

    static METHOD_GET: i32 = 1;
    static METHOD_POST: i32 = 2;
    static METHOD_PATCH: i32 = 3;
    static METHOD_DELETE: i32 = 4;

    extern {
        fn fetch_url(method: i32, url_pointer: *const u8, url_size: i32, body_pointer: *const u8, body_size: i32, ident: i32) -> i32;
    }

    pub fn get(url: &str, headers: Option<BTreeMap<&str, &str>>) -> Result<Vec<u8>, super::runnable::RunErr> {
		return do_request(METHOD_GET, url, None, headers);
	}
    
    pub fn post(url: &str, body: Option<Vec<u8>>, headers: Option<BTreeMap<&str, &str>>) -> Result<Vec<u8>, super::runnable::RunErr> {
		return do_request(METHOD_POST, url, body, headers);
	}
    
    pub fn patch(url: &str, body: Option<Vec<u8>>, headers: Option<BTreeMap<&str, &str>>) -> Result<Vec<u8>, super::runnable::RunErr> {
		return do_request(METHOD_PATCH, url, body, headers);
	}
    
    pub fn delete(url: &str, headers: Option<BTreeMap<&str, &str>>) -> Result<Vec<u8>, super::runnable::RunErr> {
		return do_request(METHOD_DELETE, url, None, headers);
	}

	fn do_request(method: i32, url: &str, body: Option<Vec<u8>>, headers: Option<BTreeMap<&str, &str>>) -> Result<Vec<u8>, super::runnable::RunErr> {
        // the URL gets encoded with headers added on the end, seperated by ::
	    // eg. https://google.com/somepage::authorization:bearer qdouwrnvgoquwnrg::anotherheader:nicetomeetyou
        let header_string = render_header_string(headers);
        
        let url_string = match header_string {
            Some(h) => format!("{}::{}", url, h),
            None => String::from(url)
        };

        let body_pointer: *const u8;
        let mut body_size: i32 = 0;

        match body {
            Some(b) => {
                let body_slice = b.as_slice();
                body_pointer = body_slice.as_ptr();
                body_size = b.len() as i32;
            },
            None => body_pointer = 0 as *const u8
        }
        
        // do the request over FFI
        let result_size = unsafe { fetch_url(method, url_string.as_str().as_ptr(), url_string.len() as i32, body_pointer, body_size, super::STATE.ident) };

        // retreive the result from the host and return it
        match super::ffi::result(result_size) {
            Ok(res) => Ok(res),
            Err(e) => {
                Err(super::runnable::RunErr::new(e.code, "failed to fetch_url"))
            }
        }
	}
	
	fn render_header_string(headers: Option<BTreeMap<&str, &str>>) -> Option<String> {
        let mut rendered: String = String::from("");
        
        let header_map = headers?;

        for key in header_map.keys() {
            rendered.push_str(key);
            rendered.push_str(":");

            let val: &str = match header_map.get(key) {
                Some(v) => v,
                None => "",
            };

            rendered.push_str(val);
            rendered.push_str("::")
        }

		return Some(String::from(rendered.trim_end_matches("::")));
	}
}

pub mod cache {
    extern {
        fn cache_set(key_pointer: *const u8, key_size: i32, value_pointer: *const u8, value_size: i32, ttl: i32, ident: i32) -> i32;
        fn cache_get(key_pointer: *const u8, key_size: i32, ident: i32) -> i32;
    }

    pub fn set(key: &str, val: Vec<u8>, ttl: i32) {
        let val_slice = val.as_slice();
        let val_ptr = val_slice.as_ptr();

        unsafe {
            cache_set(key.as_ptr(), key.len() as i32, val_ptr, val.len() as i32, ttl, super::STATE.ident);
        }
    }

    pub fn get(key: &str) -> Result<Vec<u8>, super::runnable::RunErr> {
        // do the request over FFI
        let result_size = unsafe { cache_get(key.as_ptr(), key.len() as i32, super::STATE.ident) };
        
        // retreive the result from the host and return it
        match super::ffi::result(result_size) {
            Ok(res) => Ok(res),
            Err(e) => {
                Err(super::runnable::RunErr::new(e.code, "failed to cache_get"))
            }
        }
    }
}

pub mod req {
    use super::util;

    extern {
        fn request_get_field(field_type: i32, key_pointer: *const u8, key_size: i32, ident: i32) -> i32;
    }

    static FIELD_TYPE_META: i32 = 0 as i32;
    static FIELD_TYPE_BODY: i32 = 1 as i32;
    static FIELD_TYPE_HEADER: i32 = 2 as i32;
    static FIELD_TYPE_PARAMS: i32 = 3 as i32;
    static FIELD_TYPE_STATE: i32 = 4 as i32;

    pub fn method() -> String {
        match get_field(FIELD_TYPE_META, "method") {
            Some(bytes) => return util::to_string(bytes),
            None => return String::from("")
        }
    }
    
    pub fn url() -> String {
        match get_field(FIELD_TYPE_META, "url") {
            Some(bytes) => return util::to_string(bytes),
            None => return String::from("")
        }
    }
    
    pub fn id() -> String {
        match get_field(FIELD_TYPE_META, "id") {
            Some(bytes) => return util::to_string(bytes),
            None => return String::from("")
        }
    }
    
    pub fn body_raw() -> Vec<u8> {
        match get_field(FIELD_TYPE_META, "body") {
            Some(bytes) => return bytes,
            None => return Vec::default()
        }
    }

    pub fn body_field(key: &str) -> String {
        match get_field(FIELD_TYPE_BODY, key) {
            Some(bytes) => return util::to_string(bytes),
            None => return String::from("")
        }
    }
    
    pub fn header(key: &str) -> String {
        match get_field(FIELD_TYPE_HEADER, key) {
            Some(bytes) => return util::to_string(bytes),
            None => return String::from("")
        }
    }
    
    pub fn url_param(key: &str) -> String {
        match get_field(FIELD_TYPE_PARAMS, key) {
            Some(bytes) => return util::to_string(bytes),
            None => return String::from("")
        }
    }

    pub fn state(key: &str) -> Option<String> {
        match get_field(FIELD_TYPE_STATE, key) {
            Some(bytes) => Some(util::to_string(bytes)),
            None => None
        }
    }

    pub fn state_raw(key: &str) -> Option<Vec<u8>> {
        get_field(FIELD_TYPE_STATE, key)
    }
    
    fn get_field(field_type: i32, key: &str) -> Option<Vec<u8>> {
        // make the request over FFI
        let result_size = unsafe { request_get_field(field_type, key.as_ptr(), key.len() as i32, super::STATE.ident) };

        // retreive the result from the host and return it
        match super::ffi::result(result_size) {
            Ok(res) => Some(res),
            Err(e) => {
                super::log::debug(format!("failed to request_get_field: {}", e.code).as_str());
                None
            }
        }
    }
}

pub mod resp {
    extern {
        fn resp_set_header(key_pointer: *const u8, key_size: i32, val_pointer: *const u8, val_size: i32, ident: i32);
    }

    pub fn set_header(key: &str, val: &str) {
        unsafe { resp_set_header(key.as_ptr(), key.len() as i32, val.as_ptr(), val.len() as i32, super::STATE.ident) };
    }

    pub fn content_type(ctype: &str) {
        set_header("Content-Type", ctype);
    }
}

pub mod log {
    extern {
        fn log_msg(pointer: *const u8, result_size: i32, level: i32, ident: i32);
    }
    
    pub fn debug(msg: &str) {
        log_at_level(msg, 4)
    }

    pub fn info(msg: &str) {
        log_at_level(msg, 3)
    }
    
    pub fn warn(msg: &str) {
        log_at_level(msg, 2)
    }
    
    pub fn error(msg: &str) {
        log_at_level(msg, 1)
    }

    fn log_at_level(msg: &str, level: i32) {
        unsafe { log_msg(msg.as_ptr(), msg.len() as i32, level, super::STATE.ident) };
    }
}

pub mod file {
    extern {
        fn get_static_file(name_ptr: *const u8, name_size: i32, ident: i32) -> i32;
    }

    pub fn get_static(name: &str) -> Option<Vec<u8>> {
        // do the result over FFI
        let result_size = unsafe { get_static_file(name.as_ptr(), name.len() as i32, super::STATE.ident) };

        // retreive the result from the host and return it
        match super::ffi::result(result_size) {
            Ok(res) => Some(res),
            Err(e) => {
                super::log::debug(format!("failed to get_static_file: {}", e.code).as_str());
                None
            }
        }
    }
}

pub mod util {
    pub fn to_string(input: Vec<u8>) -> String {
        String::from_utf8(input).unwrap_or_default()
    }

    pub fn to_vec(input: String) -> Vec<u8> {
        input.as_bytes().to_vec()
    }

    pub fn str_to_vec(input: &str) -> Vec<u8> {
        String::from(input).as_bytes().to_vec()
    }
}


///////////////////////////////////////
// some defaults and glue code below //
///////////////////////////////////////


// a dummy type to hold down the fort until a real Runnable is set
struct DefaultRunnable {}
impl runnable::Runnable for DefaultRunnable {
    fn run(&self, _input: Vec<u8>) -> Result<Vec<u8>, runnable::RunErr> {
        Err(runnable::RunErr::new(500, ""))
    }
}

// glue code for retreiving results of host function calls
mod ffi {
    use std::slice;
    
    extern {
        fn get_ffi_result(pointer: *const u8, ident: i32) -> i32;
    }
    
    pub fn result(size: i32) -> Result<Vec<u8>, super::runnable::RunErr> {
        // FFI functions return negative values when an error occurs
        if size < 0 {
            return Err(super::runnable::RunErr::new(size*-1, "an error was returned"));
        }

        // create some memory for the host to write into
        let mut result_mem = Vec::with_capacity(size as usize);
        let result_ptr = result_mem.as_mut_slice().as_mut_ptr() as *const u8;

        let code = unsafe {
             get_ffi_result(result_ptr, super::STATE.ident)
        };

        // check if it was successful, and then re-build the memory
        if code != 0 {
            return Err(super::runnable::RunErr::new(size*-1, "an error was returned"));
        }

        let result: &[u8] = unsafe {
            slice::from_raw_parts(result_ptr, size as usize)
        };

        Ok(Vec::from(result))
    }
}