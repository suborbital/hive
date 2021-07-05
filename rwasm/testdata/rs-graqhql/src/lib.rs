use suborbital::runnable::*;
use suborbital::graqhql::*;
use suborbital::log::*;
use suborbital::util;

struct RsGraqhql{}

impl Runnable for RsGraqhql {
    fn run(&self, input: Vec<u8>) -> Result<Vec<u8>, RunErr> {
        let in_string = String::from_utf8(input).unwrap();

        match query("https://api.rawkode.dev", "{ allProfiles { forename, surname } }") {
            Ok(response) => {
                info(util::to_string(response).as_str())
            }
            Err(e) => {
                error(e.message.as_str())
            }
        }
    
        Ok(String::from(format!("hello {}", in_string)).as_bytes().to_vec())
    }
}


// initialize the runner, do not edit below //
static RUNNABLE: &RsGraqhql = &RsGraqhql{};

#[no_mangle]
pub extern fn init() {
    use_runnable(RUNNABLE);
}
