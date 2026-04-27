pub struct Store;

impl Store {
    pub fn save(&self) {}
}

pub const VERSION: &str = "1.0.0";
static config: &str = "local";

pub fn create_repo() {}

mod nested {
    pub struct NestedStore;
}
